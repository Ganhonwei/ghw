package handler

import (
	"math"
	"strconv"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
)

// WxQuery 微信支付查询
// func WxQuery(ctos *pb.WxpayQueryReq) (stoc *pb.WxpayQueryRsp) {
// 	stoc = new(pb.WxpayQueryRsp)
// 	var transid string = ctos.GetTransid()
// 	if transid == "" {
// 		stoc.Error = pb.PayOrderError
// 		return
// 	}
// 	queryResult, err := config.Apppay.Query(transid)
// 	//glog.Infof("queryResult  %#v, err %v, transid %s", queryResult, err, transid)
// 	if err != nil {
// 		stoc.Error = pb.PayOrderError
// 		return
// 	}
// 	if queryResult.ReturnCode == "SUCCESS" &&
// 		queryResult.ResultCode == "SUCCESS" &&
// 		queryResult.TradeState == "SUCCESS" {
// 		//glog.Infof("queryResult  %#v, err %v, transid %s", queryResult, err, transid)
// 		stoc.Result = 0
// 		stoc.Orderid = queryResult.OrderId
// 	} else {
// 		stoc.Error = pb.PayOrderError
// 	}
// 	return
// }

// WxOrder 微信支付下单
// func WxOrder(ctos *pb.WxpayOrderReq, p *data.User,
// 	ip string) (stoc *pb.WxpayOrderRsp, trade string) {
// 	stoc = new(pb.WxpayOrderRsp)
// 	var waresid uint32 = ctos.GetId()
// 	var body string = ctos.GetBody()
// 	var userid string = p.GetUserid()
// 	var agent string = p.GetAgent()
// 	glog.Info("waresid ", waresid)
// 	t := wxpayOrder(waresid, userid, agent, ip, body)
// 	if t == nil {
// 		glog.Info("wx order fail:", waresid, userid)
// 		stoc.Error = pb.PayOrderFail
// 		return
// 	}
// 	payRequest := config.Apppay.NewPaymentRequest(t.Transid)
// 	payReqJSON, err := wxpay.ToJson(&payRequest)
// 	//retMap, err := wxpay.ToMap(&payRequest)
// 	if err != nil {
// 		glog.Error("wx order err:", waresid, userid, err)
// 		stoc.Error = pb.PayOrderFail
// 		return
// 	}
// 	//payReqStr := wxpay.ToXmlString(retMap)
// 	//stoc.Payreq = payReqStr
// 	trade1, err := jsoniter.Marshal(t)
// 	if err != nil {
// 		glog.Errorf("tradeRecord Marshal err %v", err)
// 		stoc.Error = pb.PayOrderFail
// 		return
// 	}
// 	trade = string(trade1)
// 	//响应
// 	stoc.Orderid = t.Id
// 	stoc.Payreq = string(payReqJSON)
// 	//glog.Info("orderid ", t.Id, " transid ", t.Transid)
// 	stoc.Id = waresid
// 	return
// }

// 下单
// func wxpayOrder(waresid uint32, userid, agent,
// 	ip, body string) *data.TradeRecord {
// 	d := config.GetShop(utils.String(waresid))
// 	if uint32(d.Payway) != data.RMB {
// 		return nil
// 	}
// 	var diamond uint32 = d.Number
// 	var price uint32 = uint32(d.Price * 100) //转换为分
// 	var itemid string = utils.String(d.Propid)
// 	//var orderid string = data.GenCporderid(userid)
// 	var orderid string = data.GenOrderid()
// 	transid, err := config.Apppay.Submit(orderid, float64(price), body, ip)
// 	glog.Debugf("orderid %s, transid %s, err %v", orderid, transid, err)
// 	if err != nil {
// 		glog.Error("wx order err:", waresid, userid, err)
// 		return nil
// 	}
// 	//transid,下单记录
// 	return &data.TradeRecord{
// 		Id:       orderid,
// 		Transid:  transid,
// 		Userid:   userid,
// 		Itemid:   itemid,
// 		Amount:   "1",
// 		Diamond:  diamond,
// 		Money:    price,
// 		Ctime:    utils.BsonNow(),
// 		DayStamp: utils.TimestampTodayTime(),
// 		Result:   data.Tradeing, //下单状态
// 		Clientip: ip,
// 		Agent:    agent,
// 	}
// }

// ActWxpayQuery 主动查询发货
// func ActWxpayQuery(orderid string) string {
// 	utils.Sleep(120) //2分钟后主动查询到账情况
// 	queryResult, err := config.Apppay.Query(orderid)
// 	if err != nil {
// 		glog.Errorf("queryResult  %#v, err %v, orderid %s", queryResult, err, orderid)
// 		return ""
// 	}
// 	return actWxpayQuery2(queryResult)
// }

// 主动查询发货
// func actWxpayQuery2(q wxpay.QueryOrderResult) string {
// 	if q.ReturnCode != "SUCCESS" ||
// 		q.ResultCode != "SUCCESS" ||
// 		q.TradeState != "SUCCESS" {
// 		glog.Errorf("actWxpayQuery2 failed %#v", q)
// 		return ""
// 	}
// 	t := new(wxpay.TradeResult)
// 	t.ReturnCode = q.ReturnCode
// 	t.ReturnMsg = q.ReturnMsg
// 	t.AppId = q.AppId
// 	t.MchId = q.MchId
// 	t.DeviceInfo = q.DeviceInfo
// 	t.NonceStr = q.NonceStr
// 	t.Sign = q.Sign
// 	t.ResultCode = q.ResultCode
// 	t.ErrCode = q.ErrCode
// 	t.ErrCodeDesc = q.ErrCodeDesc
// 	t.OpenId = q.OpenId
// 	t.IsSubscribe = q.IsSubscribe
// 	t.TradeType = q.TradeType
// 	t.BankType = q.BankType
// 	t.TotalFee = q.TotalFee
// 	t.FeeType = q.FeeType
// 	t.CashFee = q.CashFee
// 	t.CashFeeType = q.CashFeeType
// 	t.CouponFee = q.CouponFee
// 	t.CouponCount = q.CouponCount
// 	t.TransactionId = q.TransactionId
// 	t.OrderId = q.OrderId
// 	t.Attach = q.Attach
// 	t.TimeEnd = q.TimeEnd
// 	b, err := xml.Marshal(t)
// 	if err != nil {
// 		glog.Errorf("actWxpayQuery2 err %v", err)
// 		return ""
// 	}
// 	return string(b)
// }

// WxpayVerify 回调验证,gate调用
// func WxpayVerify(arg *pb.WxpayCallback) bool {
// 	result, err := wxpay.ParseTradeResult([]byte(arg.Result))
// 	if err != nil {
// 		glog.Errorf("WxpayVerify err %v, arg %#v", err, arg)
// 		return false
// 	}
// 	err = config.Apppay.RecvVerify(&result)
// 	if err != nil {
// 		glog.Errorf("recv verify %v, err:, %v", result, err)
// 		return false
// 	}
// 	return true
// }

// WxpayCallback 回调验证或主动查询发货,dbms调用
// func WxpayCallback(arg *pb.WxpayCallback) *wxpay.TradeResult {
// 	result, err := wxpay.ParseTradeResult([]byte(arg.Result))
// 	if err != nil {
// 		glog.Errorf("WxpayCallback err %v, arg %#v", err, arg)
// 		return nil
// 	}
// 	return &result
// }

// WxpayTradeVerify 发货验证
// func WxpayTradeVerify(t *wxpay.TradeResult) *data.TradeRecord {
// 	//sign
// 	// tradeRecord := &data.TradeRecord{
// 	// 	Id: t.OrderId,
// 	// 	//Transid: t.TransactionId,
// 	// }
// 	// //订单获取
// 	// tradeRecord.Get()
// 	// //glog.Infof("tradeRecord  %#v", tradeRecord)
// 	// //glog.Infof("TradeResult  %#v", t)
// 	// if tradeRecord.Transid == "" {
// 	// 	//订单不存在或其它
// 	// 	glog.Errorf("not exist orderid %v", t)
// 	// 	return nil
// 	// }
// 	// if tradeRecord.Result == 0 {
// 	// 	//重复发货
// 	// 	glog.Errorf("repeat resp %v", t)
// 	// 	return nil
// 	// }
// 	// //更新记录
// 	// tradeRecord.Transtime = t.TimeEnd
// 	// tradeRecord.Currency = t.FeeType
// 	// tradeRecord.Paytype = 403 //t.TradeType == "APP"
// 	// money, err := strconv.Atoi(t.TotalFee)
// 	// if err != nil {
// 	// 	glog.Errorf("wxpay: %v, err: %v", t, err)
// 	// }
// 	// tradeRecord.Money = uint32(money)      //转换为分
// 	// tradeRecord.Result = data.TradeSuccess //交易成功
// 	// //glog.Infof("tradeRecord  %#v", tradeRecord)
// 	// return tradeRecord
// }

// WxpaySendGoods 发货
func SendGoods(trade *data.TradeRecord, user *data.User) {
	if user == nil || user.Userid == "" {
		glog.Errorf("userid not exist trade %#v", trade)
		trade.OrderStatus = data.TradeGoods //发货失败
	}
	//update record
	if !trade.Upsert() {
		glog.Errorf("trade save failed: %#v", trade)
	}
}

// SendGoodsOnline 在线状态发货
func SendGoodsOnline(money uint32, user *data.User) {
	if user == nil {
		return
	}
	user.AddMoney(money) //TODO 在玩家进程中发消息同步
	user.UpdateMoney()
}

// SendGoodsOffline 离线状态发货
func SendGoodsOffline(money uint32, user *data.User) {
	if user == nil {
		return
	}
	user.AddMoney(money)
	user.UpdateMoney()
	if money > 0 {
		user.AddChargeTimes()
		user.UpdateChargeTimes()
	}
}

// GetGoods 获取充值货币，返回 (diamond, coin) trade.Itemid = shop.Propid , 1=钻石,2=金币
// TODO 优化关系映射
func GetGoods(trade *data.TradeRecord) (int64, int64, int64) {
	if trade.RealAmount > 0 && trade.Amount-trade.RealAmount > 200 { // 2块钱误差
		// 充值金额与订单对不上
		return int64(trade.RealAmount), 0, 0
	}
	give, _ := strconv.Atoi(trade.OtherPresent)
	diamond := int64(trade.Score) + trade.GiveCash
	withdrawal := trade.GiveWithdrawal

	// if trade.RealAmount == 0 {
	// 	return int64(trade.Score), int64(give)
	// }
	return diamond, int64(give), withdrawal
}

// FirstPay 首充
// func FirstPay(first int, user *data.User) (diamond, coin int64, msg *pb.AgentBuildUpdate) {
// 	if first != 1 {
// 		return
// 	}
// 	if user == nil {
// 		return
// 	}
// 	diamond += int64(config.GetEnv(data.ENV6))
// 	coin += int64(config.GetEnv(data.ENV7))
// 	//更新有效代理绑定
// 	msg = AgentBuildUpdateMsg(user.GetAgent(), user.GetUserid(), 0, 1, 0)
// 	return
// }

// 更新活动
func activityRecharged(online bool, trade *data.TradeRecord) {
	if online {
		msg := new(pb.ActivityRecharged)
		msg.ActId = trade.ShopId
		msg.OrderId = trade.OrderID

	}

}

// 充值之后活动变化
func PaySuccess(user *data.User, shopId, report string, shopType int32, money uint32, save bool) (rsp *pb.PaySuccessNtf) {
	var payMent config.IPay
	if shopType == data.SHOP { // 先暂时这样判断吧 TODO 优化
		payMent = GetPayMent(shopId, pb.SHOP)
	} else if shopType == data.METAL_CARD {
		payMent = GetPayMent(shopId, pb.WEEKCARD)
	} else if shopType == data.GAME_RECHARGE {
		payMent = &data.DeskRecharge{Price: money}
	} else if shopType == data.LIMITED_GIFT {
		payMent = GetPayMent(shopId, pb.LIMITEDGIFT)
	} else if shopType == data.BREAKING_GITF {
		payMent = GetPayMent(shopId, pb.BREAKING)
	} else if shopType == data.SIGN_RECHARGE {
		payMent = GetPayMent(shopId, pb.SIGN)
	} else if shopType == data.CUSTOM_RECHARGE {
		payMent = GetPayMent(shopId, pb.CUSTOM)
	} else if shopType == data.COUPON {
		payMent = GetPayMent(shopId, pb.COUPON)
	} else {
		payMent = GetPayMent(shopId, pb.ACTIVITY)
	}
	if payMent == nil {
		return
	}
	user.RechargeTarge = append(user.RechargeTarge, int(shopType))
	if save {
		user.UpdateRechargeTarge()
	}
	rsp, err := payMent.RechargeAfter(user, shopId, shopType, save)
	if err != nil {
		return
	}

	rsp.ReportId = report
	return
}

func CalWithdrawLock(user *data.User) int32 {
	if user.OutDiamond <= 0 || user.RegistArea == 2 {
		return 0
	}
	out := float64(user.OutDiamond) / 10000
	return int32(math.Round(out) * 10000)
}
