package service

import (
	"fmt"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

func (s *payService) payListArgs2M(arg args.PayListArgs, channelsMap map[string]entity.ChannelInfo) (m bson.M) {
	m = bson.M{}
	var and []bson.M
	if arg.Userid != "" {
		m["userid"] = arg.Userid
	}
	if len(arg.PayChannels) > 0 {
		m["channel_id"] = bson.M{"$in": arg.PayChannels}
	}
	if arg.OrderId != "" {
		m["_id"] = arg.OrderId
	}
	if arg.OutTradeNo != "" {
		m["out_trade_no"] = arg.OutTradeNo
	}
	ctimeM := FindByDate(arg.CtimeS, arg.CtimeE, "ctime", "ctime")
	if ctime, ok := ctimeM["ctime"]; ok {
		m["ctime"] = ctime
	}
	payTimeM := FindByDate(arg.PayTimeS, arg.PayTimeE, "pay_time", "pay_time")
	if pay_time, ok := payTimeM["pay_time"]; ok {
		m["pay_time"] = pay_time
	}
	if arg.OrderStatus != 0 {
		m["order_status"] = arg.OrderStatus
	}
	if arg.IsUTR == 1 {
		m["last_utr"] = bson.M{"$exists": true, "$ne": ""}
	} else if arg.IsUTR == 2 {
		and = append(and, bson.M{
			"$or": []bson.M{
				{"last_utr": bson.M{"$exists": false}},
				{"last_utr": bson.M{"$eq": ""}},
			},
		})
	}
	if arg.IsTag == 1 {
		m["error_msg"] = bson.M{"$exists": true, "$ne": ""}
	} else if arg.IsTag == 2 {
		and = append(and, bson.M{
			"$or": []bson.M{
				{"error_msg": bson.M{"$exists": false}},
				{"error_msg": bson.M{"$eq": ""}},
			},
		})
	}
	if arg.IsFirstPay == 1 {
		m["first_pay"] = true
	} else if arg.IsFirstPay == 2 {
		and = append(and, bson.M{
			"$or": []bson.M{
				{"first_pay": bson.M{"$exists": false}},
				{"first_pay": bson.M{"$eq": false}},
			},
		})
	}

	var classChannelIds []string
	if len(arg.ChannelClasses) > 0 {
		classMap := make(map[string][]string)
		for _, c := range channelsMap {
			classMap[c.ClassName] = append(classMap[c.ClassName], c.Name)
		}
		for _, cls := range arg.ChannelClasses {
			if pkgs, ok := classMap[cls]; ok {
				classChannelIds = append(classChannelIds, pkgs...)
			}
		}
	}

	if len(arg.ChannelIds) > 0 || len(classChannelIds) > 0 {
		m["package_id"] = bson.M{"$in": append(arg.ChannelIds, classChannelIds...)}
	}

	if arg.IsUTR == 1 || arg.IsFirstPay == 1 { // 新字段条件 || arg.IsTag == 1
		and = append([]bson.M{{"ctime": bson.M{"$gt": time.Date(2025, 3, 5, 0, 0, 0, 0, location)}}}, and...)
	}
	if len(and) > 0 {
		m["$and"] = and
	}
	fmt.Printf("%v\n", m)
	return
}

// 充值订单列表统计
func (s *payService) PayListStats(arg args.PayListArgs) (ret map[string]any, err error) {
	channelsMap := make(map[string]entity.ChannelInfo)
	if len(arg.ChannelClasses) > 0 {
		channelsMap, err = ChannelService.GetChannelsMap()
		if err != nil {
			return
		}
	}
	ret = make(map[string]any)
	m := s.payListArgs2M(arg, channelsMap)
	m["shop_type"] = bson.M{"$ne": data.WEB_TEST_SHOP}

	pipe := []bson.M{
		{"$match": m},
		{"$group": bson.M{
			"_id":        bson.M{"order_status": "$order_status", "repair": "$repair"},
			"count":      bson.M{"$sum": 1},
			"amount_sum": bson.M{"$sum": "$amount"},
		}},
	}
	stats := []bson.M{}
	err = Pays.Pipe(pipe).All(&stats)
	if err != nil {
		return
	}
	var total, payCount, autoPayCount, totalAmount, payAmount, autoPayAmount int64
	for _, stat := range stats {
		_id, ok := stat["_id"].(bson.M)
		if !ok {
			continue
		}
		order_status := int(utils.ToInt64(_id["order_status"]))
		repair, ok := _id["repair"].(bool)
		if !ok {
			repair = false
		}
		count := utils.ToInt64(stat["count"])
		amount_sum := utils.ToInt64(stat["amount_sum"])

		total += count
		totalAmount += amount_sum
		if order_status == data.TradeSuccess {
			payCount += count
			payAmount += amount_sum
			if !repair {
				autoPayCount += count
				autoPayAmount += amount_sum
			}
		}
	}
	ret["total"] = total
	ret["payCount"] = payCount
	ret["autoPayRate"] = fmt.Sprintf("%.2f%%", ComputeFloat(autoPayCount, total)*100)
	ret["totalAmount"] = Chip2Float(totalAmount)
	ret["payAmount"] = Chip2Float(payAmount)
	ret["amountAvg"] = Chip2Float(ComputeFloat(totalAmount, total))
	ret["autoPayAmountAvg"] = Chip2Float(ComputeFloat(autoPayAmount, autoPayCount))
	return
}

// 获取充值订单列表
func (s *payService) PayList(page, pageSize int, sortBy string, asc int, arg args.PayListArgs, count bool) (total int, orders []*entity.PayRecord, err error) {
	channelsMap, err := ChannelService.GetChannelsMap()
	if err != nil {
		return
	}

	skipNum, sortFieldR := parsePageAndSort(page, pageSize, sortBy, asc == 1)
	m := s.payListArgs2M(arg, channelsMap)

	// 统计总数
	if count {
		total = Count(Pays, m)
		if total == 0 {
			return
		}
	}

	err = Pays.Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&orders)
	if err != nil {
		return
	}
	err = s.payListMapping(orders, channelsMap)
	return
}

func (s *payService) payListMapping(orders []*entity.PayRecord, channelsMap map[string]entity.ChannelInfo) (err error) {
	payChannelsMap, err := GameService.GetPayChannelsMap()
	if err != nil {
		return
	}

	for _, order := range orders {
		channelId := fmt.Sprint(order.ChannelId)
		if c, ok := channelsMap[order.PackageId]; ok {
			order.FPackageAlias = c.Name1
			order.FPackageClass = c.ClassName
		}

		if c, ok := payChannelsMap[channelId]; ok {
			order.FChannel = c.Name
			order.FChannelType = c.FPtype
		}
		order.FCtime = order.Ctime.In(location).Format(utils.FORMAT)
		if !utils.SliceIn(order.OrderStatus, data.Tradeing, data.TradeOrderFail) {
			if !order.PayTime.IsZero() {
				order.FPayTime = order.PayTime.In(location).Format(utils.FORMAT)
			} else {
				order.FPayTime = order.Utime.In(location).Format(utils.FORMAT)
			}
		}

		switch order.ShopType {
		default:
			order.FShopType = "--"
		case data.SHOP:
			order.FShopType = "商城直充"
		case data.RECHARGE_GIFT:
			order.FShopType = "入门礼包"
		case data.METAL_CARD:
			order.FShopType = "周卡"
		case data.FIRST_RECHARGE:
			order.FShopType = "首充"
		case data.GAME_RECHARGE:
			order.FShopType = "局内充值"
		case data.LIMITED_GIFT:
			order.FShopType = "破冰高送"
		case data.BREAKING_GITF:
			order.FShopType = "破产礼包"
		case data.SIGN_RECHARGE:
			order.FShopType = "签到充值"
		case data.CUSTOM_RECHARGE:
			order.FShopType = "自定义充值"
		case data.Withdraw_GIFT:
			order.FShopType = "破冰解提"
		case data.COUPON:
			order.FShopType = "优惠券"
		case data.WEB_TEST_SHOP:
			order.FShopType = "后台测试单"
		}

		// 首充二充三充
		pay1ShopIds := []string{"lb1c1", "lb1c2", "lb1c3", "lb1c4", "lb1c5", "lb1c6", "lb1c7", "lb1c8"}
		pay2ShopIds := []string{"lbec1", "lbec2", "lbec3", "lbec4", "lbec5", "lbec6", "lbec7", "lbec8"}
		pay3ShopIds := []string{"lbsc1", "lbsc2", "lbsc3", "lbsc4", "lbsc5", "lbsc6", "lbsc7", "lbsc8"}
		switch true {
		case utils.SliceIn(order.ShopId, pay1ShopIds...):
			order.FShopType = "首充礼包"
		case utils.SliceIn(order.ShopId, pay2ShopIds...):
			order.FShopType = "二充礼包"
		case utils.SliceIn(order.ShopId, pay3ShopIds...):
			order.FShopType = "三充礼包"
		}

		order.FFirstPay = utils.CaseElse(order.FirstPay, "是", "否")
		order.FAmount = fmt.Sprintf("%.2f", Chip2Float(order.Amount))
		order.FRealAmount = fmt.Sprintf("%.2f", Chip2Float(order.RealAmount))
		order.FCash = fmt.Sprintf("%.2f", Chip2Float(order.Score+uint32(order.GiveCash)))
		order.FGiveCash = fmt.Sprintf("%.2f", Chip2Float(order.GiveCash))

		if order.OtherPresent != "" {
			give, err := strconv.Atoi(order.OtherPresent)
			if err != nil {
				order.FOtherPresent = "err:" + order.OtherPresent
			} else {
				order.FOtherPresent = fmt.Sprintf("%.2f", Chip2Float(give))
			}
		}

		switch order.OrderStatus {
		case data.Tradeing:
			order.FOrderStatus = "待支付"
		case data.TradeFail:
			order.FOrderStatus = "支付失败"
		case data.TradeGoods:
			order.FOrderStatus = "发货失败"
		case data.TradeSuccess:
			order.FOrderStatus = "已支付"
			if order.Repair {
				order.FOrderStatus += "(人工补单)"
			} else {
				order.FOrderStatus += "(自动上分)"
			}
		case data.TradeOrderFail:
			order.FOrderStatus = "下单失败"
		}

		for i := len(order.UserUtrs) - 1; i >= 0; i-- {
			utr := order.UserUtrs[i]
			order.FUtrs = append(order.FUtrs, fmt.Sprintf("%s: %s", time.Unix(utr.Time, 0).In(location).Format(utils.FORMAT2), utr.UTR))
		}
		// 标记信息
		if order.ErrorMsg != "" {
			order.ErrorMsgs = strings.Split(order.ErrorMsg, ";")
			utils.SliceReverse(order.ErrorMsgs)
		}

		// 补单上分信息
		if order.Repair {
			order.FRepair = fmt.Sprintf("%s %s补单上分", order.PayTime.In(location).Format(utils.FORMAT2), order.RepairAdmin)
		}
	}
	return
}

// 订单号查询订单
func (this *payService) GetPayRecord(orderid string) (record entity.PayRecord) {
	GetByQ(Pays, bson.M{"_id": orderid}, &record)
	return
}

// 更新标记状态
func (this *payService) UpdatePayTag(orderid string, errorMsg string) bool {
	m := bson.M{"_id": orderid}
	n := bson.M{"$set": bson.M{
		"error_msg": errorMsg,
	}}
	return Update(Pays, m, n)
}

// 手动补单用户信息
func (this *payService) UpdatePayRepairAdmin(orderid string, repairAdmin string) bool {
	m := bson.M{"_id": orderid}
	n := bson.M{"$set": bson.M{
		"repair_admin": repairAdmin,
	}}
	return Update(Pays, m, n)
}
