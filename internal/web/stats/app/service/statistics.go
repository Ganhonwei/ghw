package service

import (
	"errors"
	"goserver/internal/web/stats/app/entity"
	"time"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
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
	GetByQ(DataStatisticss, bson.M{"date": channel.Date, "channel": channel.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1": channel.Channel1,
			// "blogger_id":             channel.BloggerId,
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

// 新增|修改 数据汇总（新）
func (this *statisticsService) AddDataStatistics4Web(channel *entity.DataStatistics4Web) error {
	info := new(entity.DataStatistics)
	GetByQ(DataStatisticss, bson.M{"date": channel.Date, "data_types": channel.DataTypes, "channel": channel.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               channel.Channel1,
			"old_recharge_num":       channel.OldRechargeNum,
			"old_recharge_amount":    channel.OldRechargeAmount,
			"old_login_num":          channel.OldLoginNum,
			"new_register":           channel.NewRegister,
			"new_equipment":          channel.NewEquipment,
			"login_num":              channel.LoginNum,
			"next_new_register":      channel.NextNewRegister,
			"next_login":             channel.NextLogin,
			"jr_login_next_pay":      channel.JRLoginNextPay,
			"zr_pay_count":           channel.ZRPayCount,
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
			"old_recharge_num2":      channel.OldRechargeNum2,
			"new_recharge_num2":      channel.NewRechargeNum2,
			"total_recharge_num2":    channel.TotalRechargeNum2,
			"old_pay_success_order":  channel.OldPaySuccessOrder,
			"new_pay_success_order":  channel.NewPaySuccessOrder,
			"pay_login_number":       channel.PayLoginNumber,
			"pay_minutes30":          channel.PayMinutes30,
			"pay_minutes60":          channel.PayMinutes60,
			"pay_minutes120":         channel.PayMinutes120,
			"trust_user_pay":         channel.TrustUserPay,
			"trust_user_withdraw":    channel.TrustUserWithdraw,
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
func (this *statisticsService) AddUserRetained(user *entity.UserRetained4Web) error {
	info := new(entity.UserRetained4Web)
	if user.SType == 0 {
		GetByQ(UserRetaineds, bson.M{"date": user.Date, "s_type": user.SType}, info)
	} else {
		GetByQ(UserRetaineds, bson.M{"date": user.Date, "s_type": user.SType, "channel": user.Channel}, info)
	}
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"new_number": user.NewNumber,
			"s_type":     user.SType,
			"day1":       user.Day1,
			"day2":       user.Day2,
			"day3":       user.Day3,
			"day4":       user.Day4,
			"day5":       user.Day5,
			"day6":       user.Day6,
			"day7":       user.Day7,
			"day15":      user.Day15,
			"day30":      user.Day30,
			"day60":      user.Day60,
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

// 新增或更新用户留存数据
func (this *statisticsService) AddUserRetainedNew(user *entity.UserRetainedData) error {
	info := new(entity.UserRetainedData)
	GetByQ(UserRetaineds, bson.M{"date": user.Date, "channel": user.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"new_number": user.NewNumber,
			"login1":     user.Login1,
			// "register1":  user.Register1,
			"login2": user.Login2,
			// "register2":  user.Register2,
			"login3": user.Login3,
			// "register3":  user.Register3,
			"login4": user.Login4,
			// "register4":  user.Register4,
			"login5": user.Login5,
			// "register5":  user.Register5,
			"login6": user.Login6,
			// "register6":  user.Register6,
			"login7": user.Login7,
			// "register7":  user.Register7,
			"login15": user.Login15,
			// "register15": user.Register15,
			"login30": user.Login30,
			// "register30": user.Register30,
			"login60": user.Login60,
			// "register60": user.Register60,
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

func (this *statisticsService) GetUserRetainedNew(user *entity.UserRetainedData) entity.UserRetainedData {
	info := new(entity.UserRetainedData)
	GetByQ(UserRetaineds, bson.M{"date": user.Date, "channel": user.Channel}, info)
	return *info
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
func (this *statisticsService) AddPayUserRetained(user *entity.PayUserRetained4Web) error {
	info := new(entity.PayUserRetained4Web)
	if user.SType == 0 {
		GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "s_type": user.SType, "pay_user_type": user.PayUserType}, info)
	} else {
		GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "s_type": user.SType, "channel": user.Channel, "pay_user_type": user.PayUserType}, info)
	}
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"new_number": user.NewNumber,
			"s_type":     user.SType,
			"day1":       user.Day1,
			"day2":       user.Day2,
			"day3":       user.Day3,
			"day4":       user.Day4,
			"day5":       user.Day5,
			"day6":       user.Day6,
			"day7":       user.Day7,
			"day15":      user.Day15,
			"day30":      user.Day30,
			"day60":      user.Day60,
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

func (this *statisticsService) AddPayUserRetainedNew(user *entity.PayUserRetainedData) error {
	info := new(entity.PayUserRetainedData)
	GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "channel": user.Channel, "pay_user_type": user.PayUserType}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":   user.Channel1,
			"new_number": user.NewNumber,
			"pay1":       user.Pay1,
			// "register1":  user.Register1,
			"pay2": user.Pay2,
			// "register2":  user.Register2,
			"pay3": user.Pay3,
			// "register3":  user.Register3,
			"pay4": user.Pay4,
			// "register4":  user.Register4,
			"pay5": user.Pay5,
			// "register5":  user.Register5,
			"pay6": user.Pay6,
			// "register6":  user.Register6,
			"pay7": user.Pay7,
			// "register7":  user.Register7,
			"pay15": user.Pay15,
			// "register15": user.Register15,
			"pay30": user.Pay30,
			// "register30": user.Register30,
			"pay60": user.Pay60,
			// "register60": user.Register60,
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

func (this *statisticsService) GetPayUserRetainedNew(user *entity.PayUserRetainedData) entity.PayUserRetainedData {
	info := new(entity.PayUserRetainedData)
	GetByQ(PayUserRetaineds, bson.M{"date": user.Date, "channel": user.Channel, "pay_user_type": user.PayUserType}, info)
	return *info
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

// 数据汇总
func (this *statisticsService) AddDataSummary(summary *entity.DataSummary) error {
	info := new(entity.DataSummary)
	GetByQ(DataSummarys, bson.M{"date": summary.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"new_register":          summary.NewRegister,
			"new_equipment":         summary.NewEquipment,
			"login_num":             summary.LoginNum,
			"next_day_retention":    summary.NextDayRetention,
			"pay_retention":         summary.PayRetention,
			"new_recharge_num":      summary.NewRechargeNum,
			"new_recharge_amount":   summary.NewRechargeAmount,
			"new_recharge_arpu":     summary.NewRechargeArpu,
			"new_recharge_arppu":    summary.NewRechargeArppu,
			"new_user_payment_rate": summary.NewUserPaymentRate,
			"total_recharge":        summary.TotalRecharge,
			"total_amount":          summary.TotalAmount,
			"total_recharge_arpu":   summary.TotalRechargeArpu,
			"total_recharge_arppu":  summary.TotalRechargeArppu,
			"total_payment_rate":    summary.TotalPaymentRate,
			"total_withdraw_num":    summary.TotalWithdrawNum,
			"total_withdraw_amount": summary.TotalWithdrawAmount,
			"handling_charge":       summary.HandlingCharge,
			"cost_ratio":            summary.CostRatio,
		}
		if Update(DataSummarys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + summary.Id)
	} else {
		// 新增
		summary.Id = bson.NewObjectId().Hex()
		if !Insert(DataSummarys, summary) {
			return errors.New("写入失败:" + summary.Id)
		}
		return nil
	}
}

// 渠道数据
func (this *statisticsService) AddChannelData(channel *entity.ChannelData) error {
	info := new(entity.ChannelData)
	GetByQ(ChannelDatas, bson.M{"date": channel.Date, "channel": channel.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               channel.Channel1,
			"new_register":           channel.NewRegister,
			"new_equipment":          channel.NewEquipment,
			"login_num":              channel.LoginNum,
			"next_day_retention":     channel.NextDayRetention,
			"new_recharge_num":       channel.NewRechargeNum,
			"new_recharge_amount":    channel.NewRechargeAmount,
			"new_recharge_arpu":      channel.NewRechargeArpu,
			"new_recharge_arppu":     channel.NewRechargeArppu,
			"new_user_payment_rate":  channel.NewUserPaymentRate,
			"total_recharge":         channel.TotalRecharge,
			"total_amount":           channel.TotalAmount,
			"total_recharge_arpu":    channel.TotalRechargeArpu,
			"total_recharge_arppu":   channel.TotalRechargeArppu,
			"total_payment_rate":     channel.TotalPaymentRate,
			"pay_request":            channel.PayRequest,
			"pay_request_order":      channel.PayRequestOrder,
			"pay_success_order":      channel.PaySuccessOrder,
			"pay_success_rate":       channel.PaySuccessRate,
			"withdraw_request":       channel.WithdrawRequest,
			"withdraw_request_order": channel.WithdrawRequestOrder,
			"withdraw_success_order": channel.WithdrawSuccessOrder,
			"withdraw_success_rate":  channel.WithdrawSuccessRate,
			"total_withdraw_num":     channel.TotalWithdrawNum,
			"total_withdraw_amount":  channel.TotalWithdrawAmount,
			"handling_charge":        channel.HandlingCharge,
			"cost_ratio":             channel.CostRatio,
		}
		if Update(ChannelDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + channel.Id)
	} else {
		// 新增
		channel.Id = bson.NewObjectId().Hex()
		if !Insert(ChannelDatas, channel) {
			return errors.New("写入失败:" + channel.Id)
		}
		return nil
	}
}

// 实时数据
func (this *statisticsService) AddRealTimeData(res *entity.RealTimeData) error {
	info := new(entity.RealTimeData)
	// date := res.Date.Format("2006-01-02 15:04:05")
	GetByQ(RealTimes, bson.M{"date": bson.M{"$eq": res.Date}}, info)
	// targetTime := date
	// filter :=
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"login_number":    info.LoginNumber,
			"online_number":   info.OnlineNumber,
			"pay_number":      info.PayNumber,
			"pay_amount":      info.PayAmount,
			"withdraw_number": info.WithdrawNumber,
			"withdraw_amount": info.WithdrawAmount,
		}
		if Update(RealTimes, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		// 解析时间字符串
		layout := "2006-01-02 15:04:05.000"
		deleteDateStr := res.Date.Format("2006-01-02 15:04:05.000")
		deleteDate, err := time.Parse(layout, deleteDateStr)
		if err != nil {
			return errors.New("解析时间字符串错误")
		}

		// 计算删除日期
		deleteThreshold := deleteDate.Add(-3 * 24 * time.Hour)

		// 构建删除条件
		filter := bson.D{{"date", bson.D{{"$lt", deleteThreshold}}}}

		// 执行删除操作
		if !DeleteAll(RealTimes, filter) {
			return errors.New("删除之前数据失败")
		}

		res.Id = bson.NewObjectId().Hex()
		if !Insert(RealTimes, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 用户资源
func (this *statisticsService) AddUserResource(user *entity.UserResource) error {
	info := new(entity.UserResource)
	GetByQ(UserResources, bson.M{"date": user.Date, "u_type": user.UType}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"user_number": user.UserNumber,
			"diamond":     user.Diamond,
			"coin":        user.Coin,
		}
		if Update(UserResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(UserResources, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

// 新增资源流动->全局
func (this *statisticsService) AddGlobalResource(res *entity.GlobalResource) error {
	info := new(entity.GlobalResource)
	GetByQ(GlobalResources, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"login_num":    res.LoginNum,
			"new_register": res.NewRegister,
			// "total_recharge":  info.TotalRecharge,
			// "total_amount":    info.TotalAmount,
			// "withdraw_num":    info.WithdrawNum,
			// "withdraw_amount": info.WithdrawAmount,
			"put_diamond":    res.PutDiamond,
			"put_coin":       res.PutCoin,
			"expend_diamond": res.ExpendDiamond,
			"expend_coin":    res.ExpendCoin,
			"gift_diamond":   res.GiftDiamond,
			"gift_coin":      res.GiftCoin,
			// "pay_put_diamond": info.PayPutDiamond,
			// "pay_put_coin":    info.PayPutCoin,
			"expend_bonus": res.ExpendBonus,
			"put_bonus":    res.PutBonus,
		}
		if Update(GlobalResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(GlobalResources, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 新增资源流动->游戏信息
func (this *statisticsService) AddGamesResource(res *entity.GameResource) error {
	info := new(entity.GameResource)
	GetByQ(GameResources, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"tp_number":            res.TpNumber,
			"tp_diamond":           res.TpDiamond,
			"tp_coin":              res.TpCoin,
			"tp_expend_diamond":    res.TpExpendDiamond,
			"tp_expend_coin":       res.TpExpendCoin,
			"rm_number":            res.RmNumber,
			"rm_diamond":           res.RmDiamond,
			"rm_coin":              res.RmCoin,
			"rm_expend_diamond":    res.RmExpendDiamond,
			"rm_expend_coin":       res.RmExpendCoin,
			"lhd_number":           res.LhdNumber,
			"lhd_diamond":          res.LhdDiamond,
			"lhd_coin":             res.LhdCoin,
			"lhd_expend_diamond":   res.LhdExpendDiamond,
			"lhd_expend_coin":      res.LhdExpendCoin,
			"up_number":            res.UpNumber,
			"up_diamond":           res.UpDiamond,
			"up_coin":              res.UpCoin,
			"up_expend_diamond":    res.UpExpendDiamond,
			"up_expend_coin":       res.UpExpendCoin,
			"crash_number":         res.CRASHNumber,
			"crash_diamond":        res.CRASHDiamond,
			"crash_coin":           res.CRASHCoin,
			"crash_expend_diamond": res.CRASHExpendDiamond,
			"crash_expend_coin":    res.CRASHExpendCoin,
		}
		if Update(GameResources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(GameResources, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 渠道成功率
func (this *statisticsService) AddChannelRate(rate *entity.ChannelSuccessRate) error {
	info := new(entity.ChannelSuccessRate)
	GetByQ(ChannelRates, bson.M{"date": rate.Date, "package_id": rate.PackageId, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"package_name":             rate.PackageName,
			"login_number":             rate.LoginNumber,
			"pay_request":              rate.PayRequest,
			"pay_request_order":        rate.PayRequestOrder,
			"pay_success_order":        rate.PaySuccessOrder,
			"pay_success_rate":         rate.PaySuccessRate,
			"pay_success_money":        rate.PaySuccessMoney,
			"pay_fail_money":           rate.PayFailMoney,
			"withdraw_request":         rate.WithdrawRequest,
			"withdraw_request_order":   rate.WithdrawRequestOrder,
			"withdraw_success_order":   rate.WithdrawSuccessOrder,
			"withdraw_success_rate":    rate.WithdrawSuccessRate,
			"withdraw_success_money":   rate.WithdrawSuccessMoney,
			"withdraw_fail_money":      rate.WithdrawFailMoney,
			"thirdparty_order":         rate.ThirdpartyOrder,
			"thirdparty_success_order": rate.ThirdpartySuccessOrder,
			"thirdparty_order_rate":    rate.ThirdpartyOrderRate,
		}
		if Update(ChannelRates, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ChannelRates, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

func (this *statisticsService) GetLogBugFeedbackLog(m bson.M) (int64, error) {
	return int64(Count(LogBugFeedbacks, m)), nil
}

// 新增房间数据
func (this *statisticsService) AddRoomData(rate *entity.RoomData) error {
	info := new(entity.RoomData)
	GetByQ(RoomDatas, bson.M{"date": rate.Date, "channel": rate.Channel, "player_types": rate.PlayerTypes, "number_types": rate.NumberTypes}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":               rate.Channel1,
			"player_total":           rate.PlayerTotal,
			"tp_game_number":         rate.TPGameNumber,
			"tp_game_real_number":    rate.TPGameRealNumber,
			"tp_number":              rate.TPNumber,
			"tp_game_time":           rate.TPGameTime,
			"rummy_game_number":      rate.RummyGameNumber,
			"rm_game_real_number":    rate.RMGameRealNumber,
			"rummy_number":           rate.RummyNumber,
			"rummy_game_time":        rate.RummyGameTime,
			"lhd_game_number":        rate.LHDGameNumber,
			"lhd_number":             rate.LHDNumber,
			"lhd_game_time":          rate.LHDGameTime,
			"up_game_number":         rate.UPGameNumber,
			"up_number":              rate.UPNumber,
			"up_game_time":           rate.UPGameTime,
			"ak_game_number":         rate.AKGameNumber,
			"ak_game_real_number":    rate.AKGameRealNumber,
			"ak_number":              rate.AKNumber,
			"ak_game_time":           rate.AKGameTime,
			"joker_game_number":      rate.JokerGameNumber,
			"joker_game_real_number": rate.JokerGameRealNumber,
			"joker_number":           rate.JokerNumber,
			"joker_game_time":        rate.JokerGameTime,
			"crash_game_number":      rate.CrashGameNumber,
			"crash_number":           rate.CrashNumber,
			"crash_game_time":        rate.CrashGameTime,
			"ab_game_number":         rate.ABGameNumber,
			"ab_number":              rate.ABNumber,
			"ab_game_time":           rate.ABGameTime,
			"cp_game_number":         rate.CPGameNumber,
			"cp_number":              rate.CPNumber,
			"cp_game_time":           rate.CPGameTime,
			"fj_game_number":         rate.FJGameNumber,
			"fj_number":              rate.FJNumber,
			"fj_game_time":           rate.FJGameTime,
			"rb_game_number":         rate.RBGameNumber,
			"rb_number":              rate.RBNumber,
			"rb_game_time":           rate.RBGameTime,
			"rm_two_game_number":     rate.RMTwoGameNumber,
			"rm_two_number":          rate.RMTwoNumber,
			"rm_two_game_time":       rate.RMTwoGameTime,
			"tp2_game_number":        rate.TP2GameNumber,
			"tp2_number":             rate.TP2Number,
			"tp2_game_time":          rate.TP2GameTime,
			"slots_game_number":      rate.SlotsGameNumber,
			"slots_number":           rate.SlotsNumber,
			"zrsx_game_number":       rate.ZRSXGameNumber,
			"zrsx_number":            rate.ZRSXNumber,
			"tp_battle_game":         rate.TPBattleGame,
			"rm_battle_game":         rate.RMBattleGame,
			"ab_battle_game":         rate.ABBattleGame,
			"userids":                rate.Userids,
		}
		if Update(RoomDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(RoomDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 新增Bug数据
func (this *statisticsService) AddBugData(rate *entity.BugStatistics) error {
	info := new(entity.BugStatistics)
	GetByQ(BugDatas, bson.M{"date": rate.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"bug_type1": rate.BugType1,
			"bug_type2": rate.BugType2,
			"bug_type3": rate.BugType3,
			"bug_type4": rate.BugType4,
			"bug_type5": rate.BugType5,
		}
		if Update(BugDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(BugDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 新增商品购买
func (this *statisticsService) AddGoodsBuy(rate *entity.GoodsBuyData) error {
	info := new(entity.GoodsBuyData)
	GetByQ(GoodsBuys, bson.M{"date": rate.Date, "shop_type": rate.ShopType, "player_types": rate.PlayerTypes, "shop_name": rate.ShopName, "amount": rate.Amount}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"total_pull_order": rate.TotalPullOrder,
			"successful_order": rate.SuccessfulOrder,
			"pull_number":      rate.PullNumber,
			"buy_number":       rate.BuyNumber,
		}
		if Update(GoodsBuys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(GoodsBuys, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 新增|修改 AD上报统计
func (this *statisticsService) AddOrUpdateAdReport(rate *entity.AdReportData) error {
	info := new(entity.AdReportData)
	GetByQ(AdReportDatas, bson.M{"date": rate.Date, "channel": rate.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":            rate.Channel1,
			"register_count":      rate.RegisterCount,
			"register_count_day":  rate.RegisterCountDay,
			"register_count_day1": rate.RegisterCountDay1,
			"fail_register_count": rate.FailRegisterCount,
			"login_count":         rate.LoginCount,
			"login_count_day":     rate.LoginCountDay,
			"login_count_day1":    rate.LoginCountDay1,
			"fail_login_count":    rate.FailLoginCount,
			"deposit_count":       rate.DepositCount,
			"deposit_count_day":   rate.DepositCountDay,
			"deposit_count_day1":  rate.DepositCountDay1,
			"fail_deposit_count":  rate.FailDepositCount,
		}
		if Update(AdReportDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(AdReportDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 新增或更新短信统计
func (this *statisticsService) AddOrUpdateSMS(user *entity.SmsData) error {
	info := new(entity.SmsData)
	GetByQ(SmsDatas, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"total_send_sms": user.TotalSendSMS,
			"send_count":     user.SendCount,
			"use_sms":        user.UseSMS,
		}
		if Update(SmsDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(SmsDatas, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

// 新增或更新充值来源
func (this *statisticsService) AddOrUpdatePaySource(user *entity.PaySourceData) error {
	info := new(entity.PaySourceData)
	GetByQ(PaySources, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"tp_amount":     user.TPAmount,
			"tp_number":     user.TPNumber,
			"lhd_amount":    user.LHDAmount,
			"lhd_number":    user.LHDNumber,
			"rm_amount":     user.RMAmount,
			"rm_number":     user.RMNumber,
			"up_amount":     user.UPAmount,
			"up_number":     user.UPNumber,
			"ak_amount":     user.AKAmount,
			"ak_number":     user.AKNumber,
			"joker_amount":  user.JOKERAmount,
			"joker_number":  user.JOKERNumber,
			"crash_amount":  user.CRASHAmount,
			"crash_number":  user.CRASHNumber,
			"ab_amount":     user.ABAmount,
			"ab_number":     user.ABNumber,
			"cp_amount":     user.CPAmount,
			"cp_number":     user.CPNumber,
			"fj_amount":     user.FJAmount,
			"fj_number":     user.FJNumber,
			"rm_two_amount": user.RMTwoAmount,
			"rm_two_number": user.RMTwoNumber,
		}
		if Update(PaySources, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(PaySources, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

// 新增或更新对战房埋点
func (this *statisticsService) AddOrUpdateBattleRoom(user *entity.BattleRoomData) error {
	info := new(entity.BattleRoomData)
	GetByQ(BattleRooms, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"clicks_count":         user.ClicksCount,
			"total_number":         user.TotalNumber,
			"create_room_count":    user.CreateRoomCount,
			"create_room_number":   user.CreateRoomNumber,
			"success_enter_count":  user.SuccessEnterCount,
			"success_enter_number": user.SuccessEnterNumber,
			"tp_enter_count":       user.TPEnterCount,
			"tp_enter_number":      user.TPEnterNumber,
			"rm_enter_count":       user.RMEnterCount,
			"rm_enter_number":      user.RMEnterNumber,
			"ab_enter_count":       user.ABEnterCount,
			"ab_enter_number":      user.ABEnterNumber,
			"join_room_count":      user.JoinRoomCount,
			"join_room_number":     user.JoinRoomNumber,
			"success_join_count":   user.SuccessJoinCount,
			"success_join_number":  user.SuccessJoinNumber,
			"tp_join_count":        user.TPJoinCount,
			"tp_join_number":       user.TPJoinNumber,
			"rm_join_count":        user.RMJoinCount,
			"rm_join_number":       user.RMJoinNumber,
			"ab_join_count":        user.ABJoinCount,
			"ab_join_number":       user.ABJoinNumber,
		}
		if Update(BattleRooms, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(BattleRooms, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}

// 新增或更新用户游戏局数
func (this *statisticsService) AddOrUpdateUserGame(addinfo *entity.UserGameData) error {
	info := new(entity.UserGameData)
	GetByQ(UserGameDatas, bson.M{"_id": addinfo.Id}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"number":      addinfo.Number,
			"game_time":   addinfo.GameTime,
			"bets":        addinfo.Bets,
			"win_bets":    addinfo.WinBets,
			"lose_bets":   addinfo.LoseBets,
			"profit":      addinfo.Profit,
			"win_number":  addinfo.WinNumber,
			"lose_number": addinfo.LoseNumber,
		}
		if Update(UserGameDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + info.Id)
	} else {
		// 新增
		// addinfo.Id = bson.NewObjectId().Hex()
		if !Insert(UserGameDatas, addinfo) {
			return errors.New("写入失败:" + addinfo.Id)
		}
		return nil
	}
}

func (this *statisticsService) AddOrUpdateUserGameStrategy(addinfo *entity.GameStrategy) error {
	info := new(entity.GameStrategy)
	GetByQ(UserGameStrategys, bson.M{"userid": addinfo.UserId, "gtype": addinfo.Gtype, "strategy_id": addinfo.StrategyId, "ctime": addinfo.Ctime}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"enter_number":     addinfo.EnterNumber,
			"effective_number": addinfo.EffectiveNumber,
			"effective_income": addinfo.EffectiveIncome,
		}
		if Update(UserGameStrategys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + info.Id)
	} else {
		// 新增
		addinfo.Id = bson.NewObjectId().Hex()
		if !Insert(UserGameStrategys, addinfo) {
			return errors.New("写入失败:" + addinfo.Id)
		}
		return nil
	}
}

// 新增或更新拼多多统计
func (this *statisticsService) AddOrUpdatePddStat(pdd *entity.PddStat) error {
	info := new(entity.PddStat)
	GetByQ(PddStats, bson.M{"date": pdd.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"logined_users":             pdd.LoginedUsers,
			"pdd_users":                 pdd.PddUsers,
			"be_invote_users":           pdd.BeInvoteUsers,
			"be_invote_logined_users":   pdd.BeInvoteLoginedUsers,
			"draw_users":                pdd.DrawUsers,
			"be_invote_user_draws":      pdd.BeInvoteUserDraws,
			"user_draws":                pdd.UserDraws,
			"award_users":               pdd.AwardUsers,
			"award_jackpot":             pdd.AwardJackpot,
			"be_invote_reg_users":       pdd.BeInvoteRegUsers,
			"be_invote_devices":         pdd.BeInvoteDevices,
			"share2users":               pdd.Share2Users,
			"share2draw_users":          pdd.Share2DrawUsers,
			"share2devices":             pdd.Share2Devices,
			"played_reg24":              pdd.PlayedReg24,
			"reg_user_chages":           pdd.RegUserChages,
			"reg_user_charge_amount":    pdd.RegUserChargeAmount,
			"reg_user_withdraw_amount":  pdd.RegUserWithdrawAmount,
			"charge_amount":             pdd.ChargeAmount,
			"withdraw_amount":           pdd.WithdrawAmount,
			"be_invote_charge_amount":   pdd.BeInvoteChargeAmount,
			"be_invote_withdraw_amount": pdd.BeInvoteWithdrawAmount,
		}
		if Update(PddStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + pdd.Id)
	} else {
		// 新增
		pdd.Id = bson.NewObjectId().Hex()
		if !Insert(PddStats, pdd) {
			return errors.New("写入失败:" + pdd.Id)
		}
		return nil
	}
}

// 新增或更新Crash统计
func (this *statisticsService) AddOrUpdateCrashPlayerStat(stat *entity.CrashPlayerStat) error {
	info := new(entity.CrashPlayerStat)
	GetByQ(CrashPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"new_reg":                 stat.NewReg,
			"all_rounds":              stat.AllRounds,
			"game_times":              stat.GameTimes,
			"bets":                    stat.Bets,
			"bet_rounds":              stat.BetRounds,
			"win_rounds":              stat.WinRounds,
			"lose_rounds":             stat.LoseRounds,
			"tie_rounds":              stat.TieRounds,
			"win_bets":                stat.WinBets,
			"lose_bets":               stat.LoseBets,
			"tie_bets":                stat.TieBets,
			"wins":                    stat.Wins,
			"loses":                   stat.Loses,
			"cash":                    stat.Cash,
			"mulpitle_sum":            stat.MulpitleSum,
			"mulpitles":               stat.Mulpitles,
			"win_escape_mulpitle_sum": stat.WinEscapeMulpitleSum,
			"win_escape_mulpitles":    stat.WinEscapeMulpitles,
			"win_mulpitle_sum":        stat.WinMulpitleSum,
			"win_mulpitles":           stat.WinMulpitles,
			"lose_mulpitle_sum":       stat.LoseMulpitleSum,
			"lose_mulpitles":          stat.LoseMulpitles,
			"round_bet_avg":           stat.RoundBetAvg,
			"ad__bundle_id":           stat.AD_BundleId,
			"channel1":                stat.Channel1,
			"regist_area":             stat.RegistArea,
			"money":                   stat.Money,
			"observe_rounds":          stat.ObserveRounds,
		}
		if Update(CrashPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(CrashPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Plane统计
func (this *statisticsService) AddOrUpdatePlanePlayerStat(stat *entity.CrashPlayerStat) error {
	info := new(entity.CrashPlayerStat)
	GetByQ(PlanePlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"new_reg":                 stat.NewReg,
			"all_rounds":              stat.AllRounds,
			"game_times":              stat.GameTimes,
			"bets":                    stat.Bets,
			"bet_rounds":              stat.BetRounds,
			"win_rounds":              stat.WinRounds,
			"lose_rounds":             stat.LoseRounds,
			"tie_rounds":              stat.TieRounds,
			"win_bets":                stat.WinBets,
			"lose_bets":               stat.LoseBets,
			"tie_bets":                stat.TieBets,
			"wins":                    stat.Wins,
			"loses":                   stat.Loses,
			"cash":                    stat.Cash,
			"mulpitle_sum":            stat.MulpitleSum,
			"mulpitles":               stat.Mulpitles,
			"win_escape_mulpitle_sum": stat.WinEscapeMulpitleSum,
			"win_escape_mulpitles":    stat.WinEscapeMulpitles,
			"win_mulpitle_sum":        stat.WinMulpitleSum,
			"win_mulpitles":           stat.WinMulpitles,
			"lose_mulpitle_sum":       stat.LoseMulpitleSum,
			"lose_mulpitles":          stat.LoseMulpitles,
			"round_bet_avg":           stat.RoundBetAvg,
			"ad__bundle_id":           stat.AD_BundleId,
			"channel1":                stat.Channel1,
			"regist_area":             stat.RegistArea,
			"money":                   stat.Money,
			"observe_rounds":          stat.ObserveRounds,
		}
		if Update(PlanePlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(PlanePlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash扶摇直上策略统计
func (this *statisticsService) AddOrUpdateCrashFYZSStat(col *mgo.Collection, stat *entity.CrashFYZSStat) error {
	info := new(entity.CrashFYZSStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                        stat.Date,
			"date_str":                    stat.DateStr,
			"userid":                      stat.Userid,
			"is_fit":                      stat.IsFit,
			"is_trigger":                  stat.IsTrigger,
			"tigger_times":                stat.TiggerTimes,
			"all_tigger_times":            stat.AllTiggerTimes,
			"all_evo_times":               stat.AllEvoTimes,
			"win_rounds":                  stat.WinRounds,
			"lose_rounds":                 stat.LoseRounds,
			"maximum_escape_multiple":     stat.MaximumEscapeMultiple,
			"minimum_escape_multiple":     stat.MinimumEscapeMultiple,
			"average_escape_multiple":     stat.AverageEscapeMultiple,
			"win_mulpitle_sum":            stat.WinMulpitleSum,
			"bets":                        stat.Bets,
			"bet_rounds":                  stat.BetRounds,
			"maximum_escape_multiple_bet": stat.MaximumEscapeMultipleBet,
			"minimum_escape_multiple_bet": stat.MinimumEscapeMultipleBet,
			"lose_mulpitle_sum":           stat.LoseMulpitleSum,
			"win_bets":                    stat.WinBets,
			"lose_bets":                   stat.LoseBets,
			"wins":                        stat.Wins,
			"loses":                       stat.Loses,
			"cash":                        stat.Cash,
			"win_escape_multiple":         stat.WinEscapeMultiple,
			"ad__bundle_id":               stat.AD_BundleId,
			"channel1":                    stat.Channel1,
			"regist_area":                 stat.RegistArea,
			"all_rounds":                  stat.AllRounds,
			"round_bet_avg":               stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash欲薅无门策略统计
func (this *statisticsService) AddOrUpdateCrashYHWMStat(col *mgo.Collection, stat *entity.CrashYHWMStat) error {
	info := new(entity.CrashYHWMStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                stat.Date,
			"date_str":            stat.DateStr,
			"userid":              stat.Userid,
			"is_fit":              stat.IsFit,
			"is_trigger":          stat.IsTrigger,
			"tigger_times":        stat.TiggerTimes,
			"all_tigger_times":    stat.AllTiggerTimes,
			"all_evo_times":       stat.AllEvoTimes,
			"win_rounds":          stat.WinRounds,
			"lose_rounds":         stat.LoseRounds,
			"burst_number":        stat.BurstNumber,
			"burst_amount":        stat.BurstAmount,
			"total_reap_amount":   stat.TotalReapAmount,
			"win_mulpitle_sum":    stat.WinMulpitleSum,
			"bets":                stat.Bets,
			"bet_rounds":          stat.BetRounds,
			"win_bets":            stat.WinBets,
			"lose_bets":           stat.LoseBets,
			"wins":                stat.Wins,
			"loses":               stat.Loses,
			"cash":                stat.Cash,
			"win_escape_multiple": stat.WinEscapeMultiple,
			"ad__bundle_id":       stat.AD_BundleId,
			"channel1":            stat.Channel1,
			"regist_area":         stat.RegistArea,
			"all_rounds":          stat.AllRounds,
			"round_bet_avg":       stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash起死回生策略统计
func (this *statisticsService) AddOrUpdateCrashQSHSStat(col *mgo.Collection, stat *entity.CrashQSHSStat) error {
	info := new(entity.CrashQSHSStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                     stat.Date,
			"date_str":                 stat.DateStr,
			"userid":                   stat.Userid,
			"allin_number":             stat.AllinNumber,
			"allin_bets":               stat.AllinBets,
			"all_evo_times":            stat.AllEvoTimes,
			"tigger_bets":              stat.TiggerBets,
			"win_rounds":               stat.WinRounds,
			"lose_rounds":              stat.LoseRounds,
			"tigger_win_mulpitle_sum":  stat.TiggerWinMulpitleSum,
			"maximum_escape_multiple":  stat.MaximumEscapeMultiple,
			"lose_max_escape_multiple": stat.LoseMaxEscapeMultiple,
			"lose_escape_multiple":     stat.LoseEscapeMultiple,
			"bets":                     stat.Bets,
			"all_rounds":               stat.AllRounds,
			"all_win_bets":             stat.AllWinBets,
			"all_lose_bets":            stat.AllLoseBets,
			"win_bets":                 stat.WinBets,
			"lose_bets":                stat.LoseBets,
			"wins":                     stat.Wins,
			"loses":                    stat.Loses,
			"cash":                     stat.Cash,
			"withdraw_amount":          stat.WithdrawAmount,
			"pay_amount":               stat.PayAmount,
			"carry_amount":             stat.CarryAmount,
			"ad__bundle_id":            stat.AD_BundleId,
			"channel1":                 stat.Channel1,
			"regist_area":              stat.RegistArea,
			"round_bet_avg":            stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash奖池风控策略统计
func (this *statisticsService) AddOrUpdateCrashJCFKStat(col *mgo.Collection, stat *entity.CrashJCFKStat) error {
	info := new(entity.CrashJCFKStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"trigger_times":           stat.TriggerTimes,
			"trigger_amount":          stat.TriggerAmount,
			"all_evo_times":           stat.AllEvoTimes,
			"tigger_bets":             stat.TiggerBets,
			"maximum_escape_multiple": stat.MaximumEscapeMultiple,
			"tigger_mulpitle_sum":     stat.TiggerMulpitleSum,
			"win_rounds":              stat.WinRounds,
			"win_mulpitle_sum":        stat.WinMulpitleSum,
			"lose_rounds":             stat.LoseRounds,
			"lose_mulpitle_sum":       stat.LoseMulpitleSum,
			"trigger_rounds":          stat.TriggerRounds,
			"win_bets":                stat.WinBets,
			"lose_bets":               stat.LoseBets,
			"wins":                    stat.Wins,
			"loses":                   stat.Loses,
			"cash":                    stat.Cash,
			"bets":                    stat.Bets,
			"all_rounds":              stat.AllRounds,
			"all_win_bets":            stat.AllWinBets,
			"all_lose_bets":           stat.AllLoseBets,
			"withdraw_amount":         stat.WithdrawAmount,
			"pay_amount":              stat.PayAmount,
			"carry_amount":            stat.CarryAmount,
			"round_bet_avg":           stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash冒险奖励策略统计
func (this *statisticsService) AddOrUpdateCrashMXJLStat(col *mgo.Collection, stat *entity.CrashMXJLStat) error {
	info := new(entity.CrashMXJLStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"bets":                    stat.Bets,
			"all_rounds":              stat.AllRounds,
			"all_win_rounds":          stat.AllWinRounds,
			"all_win_bets":            stat.AllWinBets,
			"all_lose_bets":           stat.AllLoseBets,
			"all_evo_times":           stat.AllEvoTimes,
			"tigger_bets":             stat.TiggerBets,
			"maximum_escape_multiple": stat.MaximumEscapeMultiple,
			"win_max_mulpitle":        stat.WinMaxMulpitle,
			"tigger_mulpitle_sum":     stat.TiggerMulpitleSum,
			"win_rounds":              stat.WinRounds,
			"win_mulpitle_sum":        stat.WinMulpitleSum,
			"lose_rounds":             stat.LoseRounds,
			"lose_mulpitle_sum":       stat.LoseMulpitleSum,
			"win_bets":                stat.WinBets,
			"lose_bets":               stat.LoseBets,
			"wins":                    stat.Wins,
			"loses":                   stat.Loses,
			"cash":                    stat.Cash,
			"withdraw_amount":         stat.WithdrawAmount,
			"pay_amount":              stat.PayAmount,
			"carry_amount":            stat.CarryAmount,
			"round_bet_avg":           stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新Crash人狂有祸策略统计
func (this *statisticsService) AddOrUpdateCrashRKYSStat(col *mgo.Collection, stat *entity.CrashRKYSStat) error {
	info := new(entity.CrashMXJLStat)
	GetByQ(col, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"bets":                    stat.Bets,
			"all_rounds":              stat.AllRounds,
			"all_win_bets":            stat.AllWinBets,
			"all_lose_bets":           stat.AllLoseBets,
			"all_evo_times":           stat.AllEvoTimes,
			"tigger_bets":             stat.TiggerBets,
			"maximum_escape_multiple": stat.MaximumEscapeMultiple,
			"tigger_mulpitle_sum":     stat.TiggerMulpitleSum,
			"win_rounds":              stat.WinRounds,
			"win_mulpitle_sum":        stat.WinMulpitleSum,
			"lose_rounds":             stat.LoseRounds,
			"lose_mulpitle_sum":       stat.LoseMulpitleSum,
			"rf_rounds":               stat.RFRounds,
			"rf_bets":                 stat.RFBets,
			"rf_lose_rounds":          stat.RFLoseRounds,
			"rf_reap_bets":            stat.RFReapBets,
			"win_bets":                stat.WinBets,
			"lose_bets":               stat.LoseBets,
			"wins":                    stat.Wins,
			"loses":                   stat.Loses,
			"cash":                    stat.Cash,
			"withdraw_amount":         stat.WithdrawAmount,
			"pay_amount":              stat.PayAmount,
			"carry_amount":            stat.CarryAmount,
			"round_bet_avg":           stat.RoundBetAvg,
		}
		if Update(col, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(col, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新排行榜数据
func (this *statisticsService) AddOrUpdateRankingList(newinfo *entity.RankingList) error {
	info := new(entity.RankingList)
	GetByQ(RankingLists, bson.M{"date": newinfo.Date, "r_type": newinfo.RType, "userid": newinfo.Userid}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"no":              newinfo.No,
			"regist_area":     newinfo.RegistArea,
			"pay_amount":      newinfo.PayAmount,
			"withdraw_amount": newinfo.WithdrawAmount,
			"carry_amount":    newinfo.CarryAmount,
			"profit_amount":   newinfo.ProfitAmount,
			"e_time":          newinfo.ETime,
		}
		if Update(RankingLists, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + newinfo.Id)
	} else {
		// 新增
		newinfo.Id = bson.NewObjectId().Hex()
		if !Insert(RankingLists, newinfo) {
			return errors.New("写入失败:" + newinfo.Id)
		}
		return nil
	}
}

// TP统计
func (this *statisticsService) AddOrUpdateTpPlayerStat(stat *entity.TPPlayerStat) error {
	info := new(entity.TPPlayerStat)
	GetByQ(TPPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":        stat.Date,
			"date_str":    stat.DateStr,
			"userid":      stat.Userid,
			"new_reg":     stat.NewReg,
			"all_rounds":  stat.AllRounds,
			"game_times":  stat.GameTimes,
			"bets":        stat.Bets,
			"bet_rounds":  stat.BetRounds,
			"win_rounds":  stat.WinRounds,
			"lose_rounds": stat.LoseRounds,
			"tie_rounds":  stat.TieRounds,
			"win_bets":    stat.WinBets,
			"lose_bets":   stat.LoseBets,
			"tie_bets":    stat.TieBets,
			"wins":        stat.Wins,
			"loses":       stat.Loses,
			"cash":        stat.Cash,
			// "mulpitle_sum":            stat.MulpitleSum,
			// "mulpitles":               stat.Mulpitles,
			// "win_escape_mulpitle_sum": stat.WinEscapeMulpitleSum,
			// "win_escape_mulpitles":    stat.WinEscapeMulpitles,
			// "win_mulpitle_sum":        stat.WinMulpitleSum,
			// "win_mulpitles":           stat.WinMulpitles,
			// "lose_mulpitle_sum":       stat.LoseMulpitleSum,
			// "lose_mulpitles":          stat.LoseMulpitles,
			"round_bet_avg":  stat.RoundBetAvg,
			"ad__bundle_id":  stat.AD_BundleId,
			"channel1":       stat.Channel1,
			"regist_area":    stat.RegistArea,
			"money":          stat.Money,
			"observe_rounds": stat.ObserveRounds,
		}
		if Update(TPPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(TPPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// TP房间统计
func (this *statisticsService) AddOrUpdateTpRoomStat(stat *entity.TPRoomStat) error {
	info := new(entity.TPRoomStat)
	GetByQ(TPRoomPlayerStats, bson.M{"date": stat.Date, "userid": stat.Userid, "roomid": stat.Roomid}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                       stat.Date,
			"date_str":                   stat.DateStr,
			"userid":                     stat.Userid,
			"roomid":                     stat.Roomid,
			"all_rounds":                 stat.AllRounds,
			"game_time":                  stat.GameTime,
			"bets":                       stat.Bets,
			"win_rounds":                 stat.WinRounds,
			"lose_rounds":                stat.LoseRounds,
			"win_bets":                   stat.WinBets,
			"lose_bets":                  stat.LoseBets,
			"up_number":                  stat.UpNumber,
			"down_number":                stat.DownNumber,
			"two_rounds":                 stat.TwoRounds,
			"win_two_rounds":             stat.WinTwoRounds,
			"two_bets":                   stat.TwoBets,
			"win_two_bets":               stat.WinTwoBets,
			"three_rounds":               stat.ThreeRounds,
			"win_three_rounds":           stat.WinThreeRounds,
			"three_bets":                 stat.ThreeBets,
			"win_three_bets":             stat.WinThreeBets,
			"four_rounds":                stat.FourRounds,
			"win_four_rounds":            stat.WinFourRounds,
			"four_bets":                  stat.FourBets,
			"win_four_bets":              stat.WinFourBets,
			"five_rounds":                stat.FiveRounds,
			"win_five_rounds":            stat.WinFiveRounds,
			"five_bets":                  stat.FiveBets,
			"win_five_bets":              stat.WinFiveBets,
			"chip_pool":                  stat.ChipPool,
			"bz_up_number":               stat.BZUpNumber,
			"bz_up_rounds":               stat.BZUpRounds,
			"bz_up_bets":                 stat.BZUpBets,
			"bz_up_win_rounds":           stat.BZUpWinRounds,
			"bz_up_lose_rounds":          stat.BZUpLoseRounds,
			"bz_up_two_discard_num":      stat.BZUpTwoDiscardNum,
			"bz_up_two_follow_num":       stat.BZUpTwoFollowNum,
			"bz_up_three_discard_num":    stat.BZUpThreeDiscardNum,
			"bz_up_three_follow_num":     stat.BZUpThreeFollowNum,
			"bz_up_four_discard_num":     stat.BZUpFourDiscardNum,
			"bz_up_four_follow_num":      stat.BZUpFourFollowNum,
			"bz_up_five_discard_num":     stat.BZUpFiveDiscardNum,
			"bz_up_five_follow_num":      stat.BZUpFiveFollowNum,
			"bz_down_number":             stat.BZDownNumber,
			"bz_down_rounds":             stat.BZDownRounds,
			"bz_down_bets":               stat.BZDownBets,
			"bz_down_win_rounds":         stat.BZDownWinRounds,
			"bz_down_lose_rounds":        stat.BZDownLoseRounds,
			"bz_down_two_discard_num":    stat.BZDownTwoDiscardNum,
			"bz_down_two_follow_num":     stat.BZDownTwoFollowNum,
			"bz_down_three_discard_num":  stat.BZDownThreeDiscardNum,
			"bz_down_three_follow_num":   stat.BZDownThreeFollowNum,
			"bz_down_four_discard_num":   stat.BZDownFourDiscardNum,
			"bz_down_four_follow_num":    stat.BZDownFourFollowNum,
			"bz_down_five_discard_num":   stat.BZDownFiveDiscardNum,
			"bz_down_five_follow_num":    stat.BZDownFiveFollowNum,
			"ths_up_number":              stat.THSUpNumber,
			"ths_up_rounds":              stat.THSUpRounds,
			"ths_up_bets":                stat.THSUpBets,
			"ths_up_win_rounds":          stat.THSUpWinRounds,
			"ths_up_lose_rounds":         stat.THSUpLoseRounds,
			"ths_up_two_discard_num":     stat.THSUpTwoDiscardNum,
			"ths_up_two_follow_num":      stat.THSUpTwoFollowNum,
			"ths_up_three_discard_num":   stat.THSUpThreeDiscardNum,
			"ths_up_three_follow_num":    stat.THSUpThreeFollowNum,
			"ths_up_four_discard_num":    stat.THSUpFourDiscardNum,
			"ths_up_four_follow_num":     stat.THSUpFourFollowNum,
			"ths_up_five_discard_num":    stat.THSUpFiveDiscardNum,
			"ths_up_five_follow_num":     stat.THSUpFiveFollowNum,
			"ths_down_number":            stat.THSDownNumber,
			"ths_down_rounds":            stat.THSDownRounds,
			"ths_down_bets":              stat.THSDownBets,
			"ths_down_win_rounds":        stat.THSDownWinRounds,
			"ths_down_lose_rounds":       stat.THSDownLoseRounds,
			"ths_down_two_discard_num":   stat.THSDownTwoDiscardNum,
			"ths_down_two_follow_num":    stat.THSDownTwoFollowNum,
			"ths_down_three_discard_num": stat.THSDownThreeDiscardNum,
			"ths_down_three_follow_num":  stat.THSDownThreeFollowNum,
			"ths_down_four_discard_num":  stat.THSDownFourDiscardNum,
			"ths_down_four_follow_num":   stat.THSDownFourFollowNum,
			"ths_down_five_discard_num":  stat.THSDownFiveDiscardNum,
			"ths_down_five_follow_num":   stat.THSDownFiveFollowNum,
			"dsz_up_number":              stat.DSZUpNumber,
			"dsz_up_rounds":              stat.DSZUpRounds,
			"dsz_up_bets":                stat.DSZUpBets,
			"dsz_up_win_rounds":          stat.DSZUpWinRounds,
			"dsz_up_lose_rounds":         stat.DSZUpLoseRounds,
			"dsz_up_two_discard_num":     stat.DSZUpTwoDiscardNum,
			"dsz_up_two_follow_num":      stat.DSZUpTwoFollowNum,
			"dsz_up_three_discard_num":   stat.DSZUpThreeDiscardNum,
			"dsz_up_three_follow_num":    stat.DSZUpThreeFollowNum,
			"dsz_up_four_discard_num":    stat.DSZUpFourDiscardNum,
			"dsz_up_four_follow_num":     stat.DSZUpFourFollowNum,
			"dsz_up_five_discard_num":    stat.DSZUpFiveDiscardNum,
			"dsz_up_five_follow_num":     stat.DSZUpFiveFollowNum,
			"dsz_down_number":            stat.DSZDownNumber,
			"dsz_down_rounds":            stat.DSZDownRounds,
			"dsz_down_bets":              stat.DSZDownBets,
			"dsz_down_win_rounds":        stat.DSZDownWinRounds,
			"dsz_down_lose_rounds":       stat.DSZDownLoseRounds,
			"dsz_down_two_discard_num":   stat.DSZDownTwoDiscardNum,
			"dsz_down_two_follow_num":    stat.DSZDownTwoFollowNum,
			"dsz_down_three_discard_num": stat.DSZDownThreeDiscardNum,
			"dsz_down_three_follow_num":  stat.DSZDownThreeFollowNum,
			"dsz_down_four_discard_num":  stat.DSZDownFourDiscardNum,
			"dsz_down_four_follow_num":   stat.DSZDownFourFollowNum,
			"dsz_down_five_discard_num":  stat.DSZDownFiveDiscardNum,
			"dsz_down_five_follow_num":   stat.DSZDownFiveFollowNum,
			"dth_up_number":              stat.DTHUpNumber,
			"dth_up_rounds":              stat.DTHUpRounds,
			"dth_up_bets":                stat.DTHUpBets,
			"dth_up_win_rounds":          stat.DTHUpWinRounds,
			"dth_up_lose_rounds":         stat.DTHUpLoseRounds,
			"dth_up_two_discard_num":     stat.DTHUpTwoDiscardNum,
			"dth_up_two_follow_num":      stat.DTHUpTwoFollowNum,
			"dth_up_three_discard_num":   stat.DTHUpThreeDiscardNum,
			"dth_up_three_follow_num":    stat.DTHUpThreeFollowNum,
			"dth_up_four_discard_num":    stat.DTHUpFourDiscardNum,
			"dth_up_four_follow_num":     stat.DTHUpFourFollowNum,
			"dth_up_five_discard_num":    stat.DTHUpFiveDiscardNum,
			"dth_up_five_follow_num":     stat.DTHUpFiveFollowNum,
			"dth_down_number":            stat.DTHDownNumber,
			"dth_down_rounds":            stat.DTHDownRounds,
			"dth_down_bets":              stat.DTHDownBets,
			"dth_down_win_rounds":        stat.DTHDownWinRounds,
			"dth_down_lose_rounds":       stat.DTHDownLoseRounds,
			"dth_down_two_discard_num":   stat.DTHDownTwoDiscardNum,
			"dth_down_two_follow_num":    stat.DTHDownTwoFollowNum,
			"dth_down_three_discard_num": stat.DTHDownThreeDiscardNum,
			"dth_down_three_follow_num":  stat.DTHDownThreeFollowNum,
			"dth_down_four_discard_num":  stat.DTHDownFourDiscardNum,
			"dth_down_four_follow_num":   stat.DTHDownFourFollowNum,
			"dth_down_five_discard_num":  stat.DTHDownFiveDiscardNum,
			"dth_down_five_follow_num":   stat.DTHDownFiveFollowNum,
			"ddz_up_number":              stat.DDZUpNumber,
			"ddz_up_rounds":              stat.DDZUpRounds,
			"ddz_up_bets":                stat.DDZUpBets,
			"ddz_up_win_rounds":          stat.DDZUpWinRounds,
			"ddz_up_lose_rounds":         stat.DDZUpLoseRounds,
			"ddz_up_two_discard_num":     stat.DDZUpTwoDiscardNum,
			"ddz_up_two_follow_num":      stat.DDZUpTwoFollowNum,
			"ddz_up_three_discard_num":   stat.DDZUpThreeDiscardNum,
			"ddz_up_three_follow_num":    stat.DDZUpThreeFollowNum,
			"ddz_up_four_discard_num":    stat.DDZUpFourDiscardNum,
			"ddz_up_four_follow_num":     stat.DDZUpFourFollowNum,
			"ddz_up_five_discard_num":    stat.DDZUpFiveDiscardNum,
			"ddz_up_five_follow_num":     stat.DDZUpFiveFollowNum,
			"ddz_down_number":            stat.DDZDownNumber,
			"ddz_down_rounds":            stat.DDZDownRounds,
			"ddz_down_bets":              stat.DDZDownBets,
			"ddz_down_win_rounds":        stat.DDZDownWinRounds,
			"ddz_down_lose_rounds":       stat.DDZDownLoseRounds,
			"ddz_down_two_discard_num":   stat.DDZDownTwoDiscardNum,
			"ddz_down_two_follow_num":    stat.DDZDownTwoFollowNum,
			"ddz_down_three_discard_num": stat.DDZDownThreeDiscardNum,
			"ddz_down_three_follow_num":  stat.DDZDownThreeFollowNum,
			"ddz_down_four_discard_num":  stat.DDZDownFourDiscardNum,
			"ddz_down_four_follow_num":   stat.DDZDownFourFollowNum,
			"ddz_down_five_discard_num":  stat.DDZDownFiveDiscardNum,
			"ddz_down_five_follow_num":   stat.DDZDownFiveFollowNum,
			"dgp_up_number":              stat.DGPUpNumber,
			"dgp_up_rounds":              stat.DGPUpRounds,
			"dgp_up_bets":                stat.DGPUpBets,
			"dgp_up_win_rounds":          stat.DGPUpWinRounds,
			"dgp_up_lose_rounds":         stat.DGPUpLoseRounds,
			"dgp_up_two_discard_num":     stat.DGPUpTwoDiscardNum,
			"dgp_up_two_follow_num":      stat.DGPUpTwoFollowNum,
			"dgp_up_three_discard_num":   stat.DGPUpThreeDiscardNum,
			"dgp_up_three_follow_num":    stat.DGPUpThreeFollowNum,
			"dgp_up_four_discard_num":    stat.DGPUpFourDiscardNum,
			"dgp_up_four_follow_num":     stat.DGPUpFourFollowNum,
			"dgp_up_five_discard_num":    stat.DGPUpFiveDiscardNum,
			"dgp_up_five_follow_num":     stat.DGPUpFiveFollowNum,
			"dgp_down_number":            stat.DGPDownNumber,
			"dgp_down_rounds":            stat.DGPDownRounds,
			"dgp_down_bets":              stat.DGPDownBets,
			"dgp_down_win_rounds":        stat.DGPDownWinRounds,
			"dgp_down_lose_rounds":       stat.DGPDownLoseRounds,
			"dgp_down_two_discard_num":   stat.DGPDownTwoDiscardNum,
			"dgp_down_two_follow_num":    stat.DGPDownTwoFollowNum,
			"dgp_down_three_discard_num": stat.DGPDownThreeDiscardNum,
			"dgp_down_three_follow_num":  stat.DGPDownThreeFollowNum,
			"dgp_down_four_discard_num":  stat.DGPDownFourDiscardNum,
			"dgp_down_four_follow_num":   stat.DGPDownFourFollowNum,
			"dgp_down_five_discard_num":  stat.DGPDownFiveDiscardNum,
			"dgp_down_five_follow_num":   stat.DGPDownFiveFollowNum,
			"ad__bundle_id":              stat.AD_BundleId,
			"channel1":                   stat.Channel1,
			"regist_area":                stat.RegistArea,
			"money":                      stat.Money,
		}
		if Update(TPRoomPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(TPRoomPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新LHD统计
func (this *statisticsService) AddOrUpdateLHDPlayerStat(stat *entity.LHDPlayerStat) error {
	info := new(entity.LHDPlayerStat)
	GetByQ(LHDPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":               stat.Date,
			"date_str":           stat.DateStr,
			"userid":             stat.Userid,
			"new_reg":            stat.NewReg,
			"all_rounds":         stat.AllRounds,
			"game_times":         stat.GameTimes,
			"bets":               stat.Bets,
			"bet_rounds":         stat.BetRounds,
			"win_rounds":         stat.WinRounds,
			"lose_rounds":        stat.LoseRounds,
			"tie_rounds":         stat.TieRounds,
			"win_bets":           stat.WinBets,
			"lose_bets":          stat.LoseBets,
			"tie_bets":           stat.TieBets,
			"wins":               stat.Wins,
			"loses":              stat.Loses,
			"cash":               stat.Cash,
			"dragons":            stat.Dragons,
			"tigers":             stat.Tigers,
			"ties":               stat.Ties,
			"player_dragons":     stat.PlayerDragons,
			"player_tigers":      stat.PlayerTigers,
			"player_ties":        stat.PlayerTies,
			"player_win_dragons": stat.PlayerWinDragons,
			"player_win_tigers":  stat.PlayerWinTigers,
			"player_win_ties":    stat.PlayerWinTies,
			"player_multis":      stat.PlayerMultis,
			"round_bet_avg":      stat.RoundBetAvg,
			"ad__bundle_id":      stat.AD_BundleId,
			"channel1":           stat.Channel1,
			"regist_area":        stat.RegistArea,
			"money":              stat.Money,
			"observe_rounds":     stat.ObserveRounds,
		}
		if Update(LHDPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(LHDPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新LHD新想事成统计
func (this *statisticsService) AddOrUpdateLHDXxscStat(stat *entity.LHDXxscStat) error {
	info := new(entity.LHDPlayerStat)
	GetByQ(LHDXxscStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"strategy_players":        stat.StrategyPlayers,
			"strategy_rounds":         stat.StrategyRounds,
			"strategy_disturb_rounds": stat.StrategyDisturbRounds,
			"strategy_multi_rounds":   stat.StrategyMultiRounds,
			"strategy_bets":           stat.StrategyBets,
			"strategy_win_rounds":     stat.StrategyWinRounds,
			"strategy_wins":           stat.StrategyWins,
			"strategy_loses":          stat.StrategyLoses,
			"ad__bundle_id":           stat.AD_BundleId,
			"channel1":                stat.Channel1,
			"regist_area":             stat.RegistArea,
			"money":                   stat.Money,
		}
		if Update(LHDXxscStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(LHDXxscStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新LHD求死不能统计
func (this *statisticsService) AddOrUpdateLHDQsbnStat(stat *entity.LHDQsbnStat) error {
	info := new(entity.LHDQsbnStat)
	GetByQ(LHDQsbnStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"all_in_rounds":         stat.AllInRounds,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_loses":        stat.StrategyLoses,
			"before_back_rate":      stat.BeforeBackRate,
			"after_back_rate":       stat.AfterBackRate,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(LHDQsbnStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(LHDQsbnStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新LHD龙狂有祸统计
func (this *statisticsService) AddOrUpdateLHDLkyhStat(stat *entity.LHDLkyhStat) error {
	info := new(entity.LHDLkyhStat)
	GetByQ(LHDLkyhStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"bstimes":               stat.BSTimes,
			"bs_days":               stat.BSDays,
			"trigger_times":         stat.TriggerTimes,
			"tz":                    stat.TZ,
			"nz":                    stat.NZ,
			"bs_bets":               stat.BSBets,
			"yz_number":             stat.YZNumber,
			"yz_rounds":             stat.YZRounds,
			"yz_bets":               stat.YZBets,
			"rz":                    stat.RZ,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"bets":                  stat.Bets,
			"rounds":                stat.Rounds,
			"win_bets":              stat.WinBets,
			"lose_bets":             stat.LoseBets,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_loses":        stat.StrategyLoses,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"win_rounds":            stat.WinRounds,
			"lose_rounds":           stat.LoseRounds,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(LHDLkyhStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(LHDLkyhStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

/*
	7updown 相关统计
*/

// 新增或更新7Updown统计
func (this *statisticsService) AddOrUpdateUpdownPlayerStat(stat *entity.UpdownPlayerStat) error {
	info := new(entity.UpdownPlayerStat)
	GetByQ(UpdownPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":               stat.Date,
			"date_str":           stat.DateStr,
			"userid":             stat.Userid,
			"new_reg":            stat.NewReg,
			"all_rounds":         stat.AllRounds,
			"game_times":         stat.GameTimes,
			"bets":               stat.Bets,
			"bet_rounds":         stat.BetRounds,
			"win_rounds":         stat.WinRounds,
			"lose_rounds":        stat.LoseRounds,
			"tie_rounds":         stat.TieRounds,
			"win_bets":           stat.WinBets,
			"lose_bets":          stat.LoseBets,
			"tie_bets":           stat.TieBets,
			"wins":               stat.Wins,
			"loses":              stat.Loses,
			"cash":               stat.Cash,
			"dragons":            stat.Dragons,
			"tigers":             stat.Tigers,
			"ties":               stat.Ties,
			"player_dragons":     stat.PlayerDragons,
			"player_tigers":      stat.PlayerTigers,
			"player_ties":        stat.PlayerTies,
			"player_win_dragons": stat.PlayerWinDragons,
			"player_win_tigers":  stat.PlayerWinTigers,
			"player_win_ties":    stat.PlayerWinTies,
			"player_multis":      stat.PlayerMultis,
			"round_bet_avg":      stat.RoundBetAvg,
			"ad__bundle_id":      stat.AD_BundleId,
			"channel1":           stat.Channel1,
			"regist_area":        stat.RegistArea,
			"money":              stat.Money,
			"observe_rounds":     stat.ObserveRounds,
		}
		if Update(UpdownPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(UpdownPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新7Updown新想事成统计
func (this *statisticsService) AddOrUpdateUpdownXxscStat(stat *entity.UpdownXxscStat) error {
	info := new(entity.UpdownXxscStat)
	GetByQ(UpdownXxscStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                    stat.Date,
			"date_str":                stat.DateStr,
			"userid":                  stat.Userid,
			"strategy_players":        stat.StrategyPlayers,
			"strategy_rounds":         stat.StrategyRounds,
			"strategy_disturb_rounds": stat.StrategyDisturbRounds,
			"strategy_multi_rounds":   stat.StrategyMultiRounds,
			"strategy_bets":           stat.StrategyBets,
			"strategy_win_rounds":     stat.StrategyWinRounds,
			"strategy_wins":           stat.StrategyWins,
			"strategy_loses":          stat.StrategyLoses,
			"ad__bundle_id":           stat.AD_BundleId,
			"channel1":                stat.Channel1,
			"regist_area":             stat.RegistArea,
			"money":                   stat.Money,
		}
		if Update(UpdownXxscStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(UpdownXxscStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新7updown求死不能统计
func (this *statisticsService) AddOrUpdateUpdownQsbnStat(stat *entity.UpdownQsbnStat) error {
	info := new(entity.UpdownQsbnStat)
	GetByQ(UpdownQsbnStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"all_in_rounds":         stat.AllInRounds,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_loses":        stat.StrategyLoses,
			"before_back_rate":      stat.BeforeBackRate,
			"after_back_rate":       stat.AfterBackRate,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(UpdownQsbnStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(UpdownQsbnStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新7updown龙狂有祸统计
func (this *statisticsService) AddOrUpdateUpdownLkyhStat(stat *entity.UpdownLkyhStat) error {
	info := new(entity.UpdownLkyhStat)
	GetByQ(UpdownLkyhStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"bstimes":               stat.BSTimes,
			"bs_days":               stat.BSDays,
			"trigger_times":         stat.TriggerTimes,
			"tz":                    stat.TZ,
			"nz":                    stat.NZ,
			"bs_bets":               stat.BSBets,
			"yz_number":             stat.YZNumber,
			"yz_rounds":             stat.YZRounds,
			"yz_bets":               stat.YZBets,
			"rz":                    stat.RZ,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"bets":                  stat.Bets,
			"rounds":                stat.Rounds,
			"win_bets":              stat.WinBets,
			"lose_bets":             stat.LoseBets,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_loses":        stat.StrategyLoses,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"win_rounds":            stat.WinRounds,
			"lose_rounds":           stat.LoseRounds,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(UpdownLkyhStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(UpdownLkyhStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// 新增或更新首充分析
func (this *statisticsService) AddOrUpdateFirstCharge(stat *entity.FirstChargeAnalysis) error {
	info := new(entity.FirstChargeAnalysis)
	GetByQ(FirstCharges, bson.M{"date": stat.Date, "channel": stat.Channel}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"channel1":      stat.Channel1,
			"reg_number":    stat.RegNumber,
			"first_number":  stat.FirstNumber,
			"tp_charge":     stat.TPCharge,
			"rm_charge":     stat.RMCharge,
			"lhd_charge":    stat.LHDCharge,
			"up_charge":     stat.UPCharge,
			"ak47_charge":   stat.AK47Charge,
			"joker_charge":  stat.JokerCharge,
			"crash_charge":  stat.CrashCharge,
			"cp_charge":     stat.CPCharge,
			"ab_charge":     stat.ABCharge,
			"fj_charge":     stat.FJCharge,
			"rm_two_charge": stat.RMTwoCharge,
			"qt_number":     stat.QTNumber,
		}
		if Update(FirstCharges, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		stat.Id = bson.NewObjectId().Hex()
		// 新增
		if !Insert(FirstCharges, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

/*
AB 统计
*/
func (this *statisticsService) AddOrUpdateABPlayerStat(stat *entity.ABPlayerStat) error {
	info := new(entity.ABPlayerStat)
	GetByQ(ABPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":              stat.Date,
			"date_str":          stat.DateStr,
			"userid":            stat.Userid,
			"new_reg":           stat.NewReg,
			"all_rounds":        stat.AllRounds,
			"game_times":        stat.GameTimes,
			"bets":              stat.Bets,
			"bet_rounds":        stat.BetRounds,
			"win_rounds":        stat.WinRounds,
			"lose_rounds":       stat.LoseRounds,
			"tie_rounds":        stat.TieRounds,
			"win_bets":          stat.WinBets,
			"lose_bets":         stat.LoseBets,
			"tie_bets":          stat.TieBets,
			"wins":              stat.Wins,
			"loses":             stat.Loses,
			"cash":              stat.Cash,
			"side_winner1":      stat.SideWinner1,
			"side_winner2":      stat.SideWinner2,
			"side_winner3":      stat.SideWinner3,
			"side_winner4":      stat.SideWinner4,
			"side_winner5":      stat.SideWinner5,
			"side_winner6":      stat.SideWinner6,
			"side_winner7":      stat.SideWinner7,
			"side_winner8":      stat.SideWinner8,
			"side_winner9":      stat.SideWinner9,
			"side_winner10":     stat.SideWinner10,
			"seat_bets1":        stat.SeatBets1,
			"seat_bets2":        stat.SeatBets2,
			"seat_bets3":        stat.SeatBets3,
			"seat_bets4":        stat.SeatBets4,
			"seat_bets5":        stat.SeatBets5,
			"seat_bets6":        stat.SeatBets6,
			"seat_bets7":        stat.SeatBets7,
			"seat_bets8":        stat.SeatBets8,
			"seat_bets9":        stat.SeatBets9,
			"seat_bets10":       stat.SeatBets10,
			"player_win_seat1":  stat.PlayerWinSeat1,
			"player_win_seat2":  stat.PlayerWinSeat2,
			"player_win_seat3":  stat.PlayerWinSeat3,
			"player_win_seat4":  stat.PlayerWinSeat4,
			"player_win_seat5":  stat.PlayerWinSeat5,
			"player_win_seat6":  stat.PlayerWinSeat6,
			"player_win_seat7":  stat.PlayerWinSeat7,
			"player_win_seat8":  stat.PlayerWinSeat8,
			"player_win_seat9":  stat.PlayerWinSeat9,
			"player_win_seat10": stat.PlayerWinSeat10,
			"player_multis":     stat.PlayerMultis,
			"round_bet_avg":     stat.RoundBetAvg,
			"ad__bundle_id":     stat.AD_BundleId,
			"channel1":          stat.Channel1,
			"regist_area":       stat.RegistArea,
			"money":             stat.Money,
			"observe_rounds":    stat.ObserveRounds,
		}
		if Update(ABPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(ABPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// AB策略 安能求死
func (this *statisticsService) AddOrUpdateABAnqsStat(stat *entity.ABAnqsStat) error {
	info := new(entity.ABAnqsStat)
	GetByQ(ABAnqsStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"all_in_rounds":         stat.AllInRounds,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_wins":         stat.StrategyWins,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"strategy_loses":        stat.StrategyLoses,
			"rounds":                stat.Rounds,
			"win_rounds":            stat.WinRounds,
			"strategy_rounds":       stat.StrategyRounds,
			"wins":                  stat.Wins,
			"lose_rounds":           stat.LoseRounds,
			"loses":                 stat.Loses,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(ABAnqsStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(ABAnqsStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// AB策略 安然躺赢
func (this *statisticsService) AddOrUpdateABArtyStat(stat *entity.ABArtyStat) error {
	info := new(entity.ABArtyStat)
	GetByQ(ABArtyStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"strategy_rounds":       stat.StrategyRounds,
			"tigger_times":          stat.TiggerTimes,
			"all_evo_times":         stat.AllEvoTimes,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_loses":        stat.StrategyLoses,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(ABArtyStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(ABArtyStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

/*
CP 统计
*/
func (this *statisticsService) AddOrUpdateCPPlayerStat(stat *entity.CPPlayerStat) error {
	info := new(entity.CPPlayerStat)
	GetByQ(CPPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":             stat.Date,
			"date_str":         stat.DateStr,
			"userid":           stat.Userid,
			"new_reg":          stat.NewReg,
			"all_rounds":       stat.AllRounds,
			"game_times":       stat.GameTimes,
			"bets":             stat.Bets,
			"bet_rounds":       stat.BetRounds,
			"win_rounds":       stat.WinRounds,
			"lose_rounds":      stat.LoseRounds,
			"tie_rounds":       stat.TieRounds,
			"win_bets":         stat.WinBets,
			"lose_bets":        stat.LoseBets,
			"tie_bets":         stat.TieBets,
			"wins":             stat.Wins,
			"loses":            stat.Loses,
			"cash":             stat.Cash,
			"side_winner1":     stat.SideWinner1,
			"side_winner2":     stat.SideWinner2,
			"side_winner3":     stat.SideWinner3,
			"side_winner4":     stat.SideWinner4,
			"side_winner5":     stat.SideWinner5,
			"side_winner6":     stat.SideWinner6,
			"seat_bets1":       stat.SeatBets1,
			"seat_bets2":       stat.SeatBets2,
			"seat_bets3":       stat.SeatBets3,
			"seat_bets4":       stat.SeatBets4,
			"seat_bets5":       stat.SeatBets5,
			"seat_bets6":       stat.SeatBets6,
			"player_win_seat1": stat.PlayerWinSeat1,
			"player_win_seat2": stat.PlayerWinSeat2,
			"player_win_seat3": stat.PlayerWinSeat3,
			"player_win_seat4": stat.PlayerWinSeat4,
			"player_win_seat5": stat.PlayerWinSeat5,
			"player_win_seat6": stat.PlayerWinSeat6,
			"player_multis":    stat.PlayerMultis,
			"round_bet_avg":    stat.RoundBetAvg,
			"ad__bundle_id":    stat.AD_BundleId,
			"channel1":         stat.Channel1,
			"regist_area":      stat.RegistArea,
			"money":            stat.Money,
			"observe_rounds":   stat.ObserveRounds,
		}
		if Update(CPPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(CPPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// CP策略 来玩就赢
func (this *statisticsService) AddOrUpdateCPLwjyStat(stat *entity.CPLwjyStat) error {
	info := new(entity.CPLwjyStat)
	GetByQ(CPLwjyStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"tigger_times":          stat.TiggerTimes,
			"all_evo_times":         stat.AllEvoTimes,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_wins":         stat.StrategyWins,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"strategy_loses":        stat.StrategyLoses,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(CPLwjyStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(CPLwjyStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// CP策略 来易去难
func (this *statisticsService) AddOrUpdateCPLyqnStat(stat *entity.CPLyqnStat) error {
	info := new(entity.CPLyqnStat)
	GetByQ(CPLyqnStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_rounds":       stat.StrategyRounds,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"all_in_rounds":         stat.AllInRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_loses":        stat.StrategyLoses,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"bet_rounds":            stat.BetRounds,
			"win_rounds":            stat.WinRounds,
			"lose_rounds":           stat.LoseRounds,
			"wins":                  stat.Wins,
			"loses":                 stat.Loses,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(CPLyqnStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(CPLyqnStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

/*
RB 统计
*/
func (this *statisticsService) AddOrUpdateRBPlayerStat(stat *entity.RBPlayerStat) error {
	info := new(entity.RBPlayerStat)
	GetByQ(RBPlayerStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":             stat.Date,
			"date_str":         stat.DateStr,
			"userid":           stat.Userid,
			"new_reg":          stat.NewReg,
			"all_rounds":       stat.AllRounds,
			"game_times":       stat.GameTimes,
			"bets":             stat.Bets,
			"bet_rounds":       stat.BetRounds,
			"win_rounds":       stat.WinRounds,
			"lose_rounds":      stat.LoseRounds,
			"tie_rounds":       stat.TieRounds,
			"win_bets":         stat.WinBets,
			"lose_bets":        stat.LoseBets,
			"tie_bets":         stat.TieBets,
			"wins":             stat.Wins,
			"loses":            stat.Loses,
			"cash":             stat.Cash,
			"side_winner1":     stat.SideWinner1,
			"side_winner2":     stat.SideWinner2,
			"side_winner3":     stat.SideWinner3,
			"side_winner4":     stat.SideWinner4,
			"side_winner5":     stat.SideWinner5,
			"side_winner6":     stat.SideWinner6,
			"side_winner7":     stat.SideWinner7,
			"side_winner8":     stat.SideWinner8,
			"seat_bets1":       stat.SeatBets1,
			"seat_bets2":       stat.SeatBets2,
			"seat_bets3":       stat.SeatBets3,
			"player_win_seat1": stat.PlayerWinSeat1,
			"player_win_seat2": stat.PlayerWinSeat2,
			"player_win_seat3": stat.PlayerWinSeat3,
			"player_win_seat4": stat.PlayerWinSeat4,
			"player_win_seat5": stat.PlayerWinSeat5,
			"player_win_seat6": stat.PlayerWinSeat6,
			"player_win_seat7": stat.PlayerWinSeat7,
			"player_win_seat8": stat.PlayerWinSeat8,
			"player_multis":    stat.PlayerMultis,
			"round_bet_avg":    stat.RoundBetAvg,
			"ad__bundle_id":    stat.AD_BundleId,
			"channel1":         stat.Channel1,
			"regist_area":      stat.RegistArea,
			"money":            stat.Money,
			"observe_rounds":   stat.ObserveRounds,
		}
		if Update(RBPlayerStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(RBPlayerStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// RB策略 红运
func (this *statisticsService) AddOrUpdateRBHydtStat(stat *entity.RBHydtStat) error {
	info := new(entity.RBHydtStat)
	GetByQ(RBHydtStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_times":        stat.StrategyTimes,
			"tigger_times":          stat.TiggerTimes,
			"all_evo_times":         stat.AllEvoTimes,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_wins":         stat.StrategyWins,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"strategy_loses":        stat.StrategyLoses,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(RBHydtStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(RBHydtStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}

// RB策略 绝处
func (this *statisticsService) AddOrUpdateRBJcfsStat(stat *entity.RBJcfsStat) error {
	info := new(entity.RBJcfsStat)
	GetByQ(RBJcfsStats, bson.M{"_id": stat.Id}, info)
	if info.Userid != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"date":                  stat.Date,
			"date_str":              stat.DateStr,
			"userid":                stat.Userid,
			"strategy_players":      stat.StrategyPlayers,
			"strategy_rounds":       stat.StrategyRounds,
			"strategy_multi_rounds": stat.StrategyMultiRounds,
			"all_in_rounds":         stat.AllInRounds,
			"strategy_bets":         stat.StrategyBets,
			"strategy_wins":         stat.StrategyWins,
			"strategy_win_rounds":   stat.StrategyWinRounds,
			"strategy_loses":        stat.StrategyLoses,
			"strategy_lose_rounds":  stat.StrategyLoseRounds,
			"bet_rounds":            stat.BetRounds,
			"win_rounds":            stat.WinRounds,
			"lose_rounds":           stat.LoseRounds,
			"wins":                  stat.Wins,
			"loses":                 stat.Loses,
			"ad__bundle_id":         stat.AD_BundleId,
			"channel1":              stat.Channel1,
			"regist_area":           stat.RegistArea,
			"money":                 stat.Money,
		}
		if Update(RBJcfsStats, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + stat.Id)
	} else {
		// 新增
		if !Insert(RBJcfsStats, stat) {
			return errors.New("写入失败:" + stat.Id)
		}
		return nil
	}
}
