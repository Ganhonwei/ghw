package handler

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/table"
)

// Buy 购买
// func Buy(ctos *pb.BuyReq, p *data.User) (stoc *pb.BuyRsp,
// 	diamond, coin int64) {
// 	stoc = new(pb.BuyRsp)
// 	id := ctos.GetId()
// 	d := config.GetShop(utils.String(id))
// 	switch uint32(d.Payway) {
// 	case data.DIA:
// 		if p.GetDiamond() >= int64(d.Price) {
// 			stoc.Result = 0
// 			diamond = -1 * int64(d.Price)
// 			coin = int64(d.Number)
// 		} else {
// 			stoc.Result = 1
// 			stoc.Error = pb.NotEnoughDiamond
// 		}
// 	default:
// 		stoc.Error = pb.PayOrderFail
// 	}
// 	return
// }

// Shop 商城列表
func Shop(p *data.User) (stoc *pb.ShopRsp) {
	stoc = new(pb.ShopRsp)
	if p == nil {
		return
	}

	custom := table.GetTables().CustomShopTable.Get()
	shops := table.GetTables().ShopTable.GetDataList()
	if shops == nil {
		return
	}
	utrGive := table.GetTables().PaymentTable.Get().UtrBackRewordRate[p.RegistArea]
	if utrGive > 0 {
		stoc.UtrGive = fmt.Sprintf("%g%%", float64(utrGive)/10000*100)
	}

	// 判断新玩家拉单限制
	newerPayTimesLimit := table.GetTables().PaymentTable.Get().NewerPayTimesLimit[p.RegistArea]
	newerPaysLimit := table.GetTables().PaymentTable.Get().NewerPaysLimit[p.RegistArea]

	// stoc.MinRecharge = uint32(custom.Price[0])
	// stoc.MaxRecharge = uint32(custom.Price[1])
	for _, v := range custom.GiveRate[p.RegistArea].Nums {
		stoc.GiveRate += v
	}
	// stoc.GiveRate = custom.GiveRate[p.RegistArea]
	for _, v := range shops {
		if v.Show[p.RegistArea] == 0 {
			continue
		}
		// 新手限制
		if newerPayTimesLimit > 0 && newerPaysLimit > 0 {
			if p.ChargeTimes <= newerPayTimesLimit {
				if v.Price > newerPaysLimit {
					// glog.Infof("newer user pay success times limit: %s, pay_success_times=%d, order_amount=%d", rs.Userid, pay_success_times, arg.Amount)
					continue
				}
			}
		}
		var giveRatio int32
		for _, v := range v.GiveRatio[p.RegistArea].Nums {
			giveRatio += v
		}
		giveNumber := int64(v.Price) * int64(giveRatio) / 10000
		s := &pb.Shop{
			Id:         v.Id,            //购买
			Number:     uint32(v.Price), //兑换的数量
			Price:      uint32(v.Price), //支付价格
			GiveNumber: int32(giveNumber),
		}
		stoc.List = append(stoc.List, s)
	}
	return
}

// Order2Record 下单记录
// func Order2Record(arg *pb.TradeOrder) (msg *data.TradeRecord) {
// 	msg = &data.TradeRecord{
// 		Id:       arg.Orderid,
// 		Userid:   arg.Userid,
// 		Amount:   arg.Amount,
// 		Itemid:   arg.Itemid,
// 		Diamond:  arg.Diamond,
// 		Money:    arg.Money,
// 		Result:   int(arg.Result),
// 		Clientip: arg.Clientip,
// 	}
// 	return
// }

// JtpayTradeVerify 发货验证
// func JtpayTradeVerify(t *jtpay.NotifyResult) *data.TradeRecord {
// 	//sign
// 	tradeRecord := &data.TradeRecord{
// 		Id: t.P2_ordernumber,
// 		//Transid: t.TransactionId,
// 	}
// 	//订单获取
// 	tradeRecord.Get()
// 	//glog.Infof("tradeRecord  %#v", tradeRecord)
// 	//glog.Infof("TradeResult  %#v", t)
// 	if tradeRecord.Userid == "" {
// 		//订单不存在或其它
// 		glog.Errorf("not exist orderid %#v", t)
// 		return nil
// 	}
// 	if tradeRecord.Result == 0 {
// 		//重复发货
// 		glog.Errorf("repeat resp %#v", t)
// 		return nil
// 	}
// 	//更新记录
// 	tradeRecord.Transid = t.P5_orderid
// 	tradeRecord.Transtime = utils.Time2Str(utils.LocalTime())
// 	//tradeRecord.Paytype = t.P6_productcode
// 	//money, err := strconv.Atoi(t.P13_zfmoney)
// 	//if err != nil {
// 	//	glog.Errorf("jtpay: %v, err: %#v", t, err)
// 	//}
// 	//if uint32(money*100) != tradeRecord.Money {
// 	//	glog.Errorf("jtpay money : %#v, err: %v", t, err)
// 	//}
// 	//tradeRecord.Money = uint32(money)      //转换为分
// 	tradeRecord.Result = data.TradeSuccess //交易成功
// 	//glog.Infof("tradeRecord  %#v", tradeRecord)
// 	return tradeRecord
// }
