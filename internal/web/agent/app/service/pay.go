package service

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/agent/app/entity"
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type payService struct{}

// 获取充值订单列表
func (this *payService) PayList(page, pageSize int, m bson.M) ([]entity.PayOrder, error) {
	var list []entity.PayOrder
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	Pays.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList1(list)
	return list, nil
}

// 获取充值订单总数
func (this *payService) GetPayTotal(m bson.M) (int64, error) {
	return int64(Count(Pays, m)), nil
}

// 根据条件查询对应的支付订单
func (this *payService) GetByPayUser(m bson.M) ([]entity.PayOrder, error) {
	var list []entity.PayOrder
	err := Pays.
		Find(m).All(&list)

	return list, err
}

// 转换为分展示
func (this *payService) chipList1(list []entity.PayOrder) []entity.PayOrder {
	for k, v := range list {
		v.FAmount = Chip2Float(int64(v.Amount))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		c, _ = ConvertToIndiaTime(v.Utime.Unix())
		v.Utime = c
		c, _ = ConvertToIndiaTime(v.PayTime.Unix())
		v.PayTime = c
		list[k] = v
	}
	return list
}

// 获取提现订单列表
func (this *payService) WithdrawList(page, pageSize int, m bson.M) ([]entity.WithdrawOrder, error) {
	var list []entity.WithdrawOrder
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
	list = this.chipList2(list)
	return list, nil
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

// 转换为分展示
func (this *payService) chipList2(list []entity.WithdrawOrder) []entity.WithdrawOrder {
	userids := make([]string, 0)
	for k, v := range list {
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

/*
Author:CC
Title:提现记录 -> 自动审核提现订单
*/
func (this *payService) AutomationAudit(timestamp int64) {
	//根据时间获取待审核的提现订单
	tim := utils.Stamp2Time(timestamp)
	endTime := tim                       //.Format("2006-01-02 15:04:05")
	startTime := tim.Add(-1 * time.Hour) //.Format("2006-01-02 15:04:05")
	m := bson.M{}
	m["ctime"] = bson.M{"$gte": startTime, "$lt": endTime}
	m["order_status"] = 1
	m["is_tag"] = bson.M{"$ne": true}
	list, err := this.GetByWithdrawUser(m)
	if err != nil {
		errors.New("查询提现订单失败")
	}
	if len(list) > 0 {
		isflag := false
		for _, item := range list {
			orderid := item.OrderID
			// 检测提现玩家的设备码
			pipeline := []bson.M{
				// 根据 userid 查询用户的设备码
				bson.M{"$match": bson.M{"_id": item.Userid}},
				// 查找具有相同设备码的其他用户
				bson.M{"$lookup": bson.M{
					"from":         "col_user",    // 查询的集合名
					"localField":   "ad__adid",    // 本地集合中用于关联的字段
					"foreignField": "ad__adid",    // 目标集合中用于关联的字段
					"as":           "other_users", // 查询结果保存到的字段名
				}},
				// 过滤掉本人
				bson.M{"$project": bson.M{
					"other_users": bson.M{
						"$filter": bson.M{
							"input": "$other_users",
							"as":    "other_user",
							"cond":  bson.M{"$ne": []interface{}{"$$other_user._id", item.Userid}},
						},
					},
				}},
			}
			result := []bson.M{}
			pipe := PlayerUsers.Pipe(pipeline)
			err := pipe.All(&result)
			if err != nil {
				beego.Error("AutomationAudit fail err: ", err)
			}
			if len(result) > 0 {
				temp := result[0]["other_users"].([]interface{})
				if len(temp) > 1 {
					rerr := OrderTag(orderid, "设备码重复")
					if rerr != nil {
						beego.Error("OrderTag fail err: ", rerr)
					}
					continue
				}
			}
			// 检测提现玩家的银行卡号是否和其他玩家相同
			m3 := []bson.M{
				{
					"$match": bson.M{
						"blank_number": item.BlankNumber,
						"userid":       bson.M{"$ne": item.Userid},
					},
				},
				{
					"$group": bson.M{
						"_id": "$userid",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result3 := []bson.M{}
			pipe3 := Withdraws.Pipe(m3)
			err = pipe3.All(&result3)
			if err != nil {
				beego.Error("AutomationAudit fail err: ", err)
			}
			if len(result3) > 1 {
				rerr := OrderTag(orderid, "银行卡号重复")
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}

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

			// 检测提现玩家是否只玩了百人场。
			pipeline1 := []bson.M{
				{
					"$match": bson.M{
						// "players": bson.RegEx{Pattern: item.Userid, Options: "i"},
						"players": bson.M{
							"$regex":   fmt.Sprintf("\\b%s\\b", item.Userid),
							"$options": "i",
						},
						"gtype": bson.M{"$nin": []int{2, 3, 7}},
					},
				},
				{
					"$group": bson.M{
						"_id": "$gtype",
						"num": bson.M{
							"$sum": 1,
						},
					},
				},
			}
			result1 := []bson.M{}
			pipe1 := Details.Pipe(pipeline1)
			err = pipe1.All(&result1)
			if err != nil {
				beego.Error("AutomationAudit fail err: ", err)
			}
			if len(result1) == 0 {
				rerr := OrderTag(orderid, "只玩百人场")
				if rerr != nil {
					beego.Error("OrderTag fail err: ", rerr)
				}
				continue
			}
			// 检测提现玩家申请提现前近200局是否有相同玩家同场玩5局。（局数过多或早期对局，会让数据量太大）
			pipeline2 := []bson.M{
				{
					"$match": bson.M{
						"players": bson.M{
							"$regex":   fmt.Sprintf("\\b%s\\b", item.Userid),
							"$options": "i",
						},
						"gtype": bson.M{"$ne": 7},
						// "begin_time": bson.M{"$exists": true},
					},
				},
				{"$sort": bson.M{"begin_time": -1}},
				{"$limit": 200},
				{
					"$group": bson.M{
						"_id":     "$_id",
						"players": bson.M{"$addToSet": "$players"}, // 为每个房间创建玩家列表
						// "playerIds": bson.M{"$push": "$players"},     // 保存原始的玩家列表
					},
				},
			}
			result2 := []bson.M{}
			pipe2 := Details.Pipe(pipeline2)
			err = pipe2.All(&result2)
			if err != nil {
				beego.Error("AutomationAudit fail err: ", err)
			}
			if len(result2) > 0 {
				counts := make(map[string]int)
				isCheck := false
				for _, game := range result2 {
					temp := game["players"].([]interface{})
					strid := temp[0].(string)
					if strid != "" {
						idarr := strings.Split(strid, ",")
						for _, id := range idarr {
							if id != item.Userid {
								if len(id) <= 7 {
									counts[id]++
								}

							}
						}
					}
				}
				if len(counts) > 0 {
					for _, v := range counts {
						if v >= 5 {
							isCheck = true
						}
					}
					if isCheck {
						rerr := OrderTag(orderid, "同场五局")
						if rerr != nil {
							beego.Error("OrderTag fail err: ", rerr)
						}
						continue
					}
				}
			}

			// 检测提现玩家设备分是否≥100分。
			// 通过用户ID去查询用户设备号
			ulist, _ := PlayerService.GetUser(item.Userid)
			if ulist != nil {
				ads := make([]string, 0)
				ads = append(ads, ulist.AD_ADID)
				// 加载时获取安卓评分
				result, _ := AdGetRequest(ads)
				beego.Trace("result: ", result)
				isAndroid := true
				if result != nil {
					if len(result.Data) > 0 {
						for _, item1 := range result.Data {
							if item.Userid == ulist.Userid && item1.Ad_Id == ulist.AD_ADID && item1.Status == 0 {
								if item1.Score < 100 {
									// 评分小于100
									isAndroid = false
									beego.Trace("安卓评分: ", item1.Score)
								}
							}
						}
					} else {
						isAndroid = true
					}
				} else {
					isAndroid = true
				}
				if isAndroid {
					rerr := OrderTag(orderid, "安卓评分不符合")
					if rerr != nil {
						beego.Error("OrderTag fail err: ", rerr)
					}
					continue
				}
			}
			// 根据用户Id,订单ID 查询充值金额及提现金额
			isflag = CheckRechargeAmount(item.Userid, int(item.Amount))
			if isflag {
				// 修改该订单状态、调用提现审核接口
				aerr := OrderAudio(orderid)
				if aerr != nil {
					beego.Error("OrderAudio fail err: ", aerr)
				}
				continue
			}

			// 调用审核通过结果
			if !isflag {
				// 修改该订单状态、调用提现审核接口
				aerr := OrderAudio(orderid)
				if aerr != nil {
					beego.Error("OrderAudio fail err: ", aerr)
				}
			}

		}

	} else {
		fmt.Println("未查询到可审核的提现订单！")
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
		// ActionService.Add("withdraw_audit", name,
		// 	"", utils.String(orderid), utils.String(orderid))
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
	if err != nil {
		return err
	} else {
		// ActionService.Add("withdraw_tag", name,
		// 	"", utils.String(orderid), utils.String(orderid))
	}
	return err
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
				"order_status": bson.M{"$in": []int{1, 2, 6}},
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
	// 在总充值金额上增加105，因为首次提现105不受限制
	PayAmount += 10500
	if PayAmount > total {
		isflag = true
	}
	return isflag
}
