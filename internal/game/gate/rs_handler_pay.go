package gate

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

func (rs *RoleActor) ApplePayReq(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.ApplePayReq)
	// glog.Debugf("ApplePayReq %#v", arg)
	// rs.applePay(arg)
}

// func (rs *RoleActor) WxpayOrderReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.WxpayOrderReq)
// 	glog.Debugf("WxpayOrderReq %#v", arg)
// 	// rs.wxPay(arg)
// }

// func (rs *RoleActor) WxpayQueryReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.WxpayQueryReq)
// 	glog.Debugf("WxpayQueryReq %#v", arg)
// 	rsp := handler.WxQuery(arg)
// 	rs.Send(rsp)
// }

func (rs *RoleActor) PayGoods(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PayGoods)
	//发货
	glog.Debugf("PayGoods: %v", arg)

	// if rs.VBBank < arg.Coin {
	// 	// vb银行不够了就不送
	// 	arg.Coin = 0
	// }
	// vip
	ntf := event.Event(rs.User, event.VIP, &event.VIPEvent{Amount: int64(arg.Money)})
	if ntf != nil {
		rs.Send(ntf)
	}
	// vip等级奖励检查
	rs.vipUpgradeReward()

	// 解锁提现
	if rs.Money <= 0 {
		// rs.WithdrawLock = handler.CalWithdrawLock(rs.User)
		rs.FirstChargeVal = int32(arg.Money)
		rs.FirstChargeTime = time.Now().UnixMilli()
	} else {
		rs.LastChargeTime = time.Now().UnixMilli()
		rs.TpUserGameTime = 0
	}
	// C类充的是可提现金
	// out := handler.GetChargeWithdrawable(rs.User, arg.Diamond, arg.Money)
	//获取ltype
	ltype := handler.GetRechargeLtype(arg.ShopId, arg.ShopType)
	rs.sendGood(arg.Diamond, 0, arg.Coin, arg.Withdrawal, arg.Money, 0, ltype, arg.ShopName)
	// 同步游戏节点
	if rs.gamePid != nil {
		rs.gamePid.Tell(&pb.ChangeCurrency{Diamond: arg.Diamond, Userid: rs.Userid, Give: arg.Coin, Out: arg.Withdrawal, Money: int64(arg.Money)})
	}
	if rs.cpPid != nil {
		rs.cpPid.Tell(&pb.ChangeCurrency{Diamond: arg.Diamond, Userid: rs.Userid, Give: arg.Coin, Out: arg.Withdrawal, Money: int64(arg.Money)})
	}

	// 充值成功之后
	rs.paySuccess(arg, ctx)

	//上报AF充值
	af := &pb.AfParam{
		AppId:       rs.AppId,
		AFId:        rs.AFId,
		OS:          rs.OS,
		BundleId:    rs.BundleId,
		RefGameId:   rs.RefGameId,
		RefPkgName:  rs.RefPkgName,
		MediaSource: rs.MediaSource,
		AFKey:       rs.AFKey,
	}
	ad := &pb.ADParam{
		AdAdid:       rs.AD_ADID,
		AdKey:        rs.AD_Key,
		AdS2SCode:    rs.AD_S2S_Code,
		AdEventCode1: rs.AD_Event_Code1,
		AdEventCode2: rs.AD_Event_Code2,
		AdEventCode3: rs.AD_Event_Code3,
		AdEventCode4: rs.AD_Event_Code4,
		UserAgent:    rs.AD_User_Agent,
		BundleId:     rs.AD_BundleId,
	}
	msg1 := login.ReportDepositMsg(af, ad, float64(arg.Money)/float64(100), rs.User)
	if msg1 != nil {
		rs.adQueue = append(rs.adQueue, msg1)
		// if rs.adid == "" {
		// } else {
		// 	if msg1.AdInfo != nil {
		// 		msg1.AdInfo.ADID = rs.adid
		// 	}
		// 	rs.reportPid.Tell(msg1)
		// }
	}
}

func (rs *RoleActor) PayCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PayCurrency)
	glog.Debugf("PayCurrency %#v", arg)
	//后台或充值同步到game房间
	msg2 := handler.Pay2ChangeCurr(arg)
	if rs.gamePid != nil {
		rs.gamePid.Tell(msg2)
	}
	if rs.cpPid != nil {
		rs.cpPid.Tell(msg2)
	}
	diamond := arg.Diamond
	coin := arg.Coin
	chip := arg.Chip
	card := arg.Card
	ltype := arg.Type
	give := arg.Give
	rs.addCurrency(diamond, coin, card, chip, give, 0, 0, ltype, arg.Desc, "")
	rs.AddMoney(uint32(diamond))

	// vip
	ntf := event.Event(rs.User, event.VIP, &event.VIPEvent{Amount: arg.Diamond})
	if ntf != nil {
		rs.Send(ntf)
	}
	// vip等级奖励检查
	rs.vipUpgradeReward()

	// 邮件
	if config.SettingIsOpen(4, data.MAILREPLY) && !rs.Robot && arg.Diamond != 0 {
		ntf := event.Event(rs.User, event.GIVE_CASH, &event.GiveCash{Amount: uint32(arg.Diamond)})
		if ntf != nil {
			rs.Send(ntf)
		}
		rs.status = true
	}
	rs.status = true
}

func (rs *RoleActor) ModifyCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ModifyCurrency)
	glog.Debugf("ModifyCurrency %#v", arg)
	//后台修改同步到game房间
	msg2 := handler.ModifyChangeCurr(rs.User, arg)
	if rs.gamePid != nil {
		rs.gamePid.Tell(msg2)
	}
	if rs.cpPid != nil {
		rs.cpPid.Tell(msg2)
	}
	diamond := arg.Diamond
	give := arg.Give
	ltype := arg.Type
	rs.modifyCurrency(diamond, give, ltype)
}

func (rs *RoleActor) ExternalModifyCurrencyReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ExternalModifyCurrencyReq)
	glog.Debugf("ExternalModifyCurrencyReq %#v", arg)

	if rs.gamePid != nil { //如果在游戏中不能撤单，避免数据混乱
		rsp := &pb.ExternalModifyCurrencyRsp{Err: "already in the game"}
		ctx.Respond(rsp)
		return
	}

	// rs.externalModifyCurrency(arg.Diamond, arg.Type)
	rs.addCurrency(arg.Diamond, 0, 0, 0, 0, 0, 0, arg.Type, "", "")

	rsp := &pb.ExternalModifyCurrencyRsp{}
	rsp.Balance = rs.User.Diamond
	ctx.Respond(rsp)
}

func (rs *RoleActor) ChangeCurrency(ctx actor.Context) {
	msg := ctx.Message()
	//货币变更
	arg := msg.(*pb.ChangeCurrency)
	diamond := arg.Diamond
	coin := arg.Coin
	chip := arg.Chip
	card := arg.Card
	ltype := arg.Type
	rs.addCurrency(diamond, coin, card, chip, arg.Give, arg.Out, arg.Priv, ltype, arg.Desc, arg.WaterId)
	if !arg.SyncCpOff && rs.cpPid != nil {
		rs.cpPid.Tell(arg)
	}
	if arg.SyncGame && rs.gamePid != nil {
		rs.gamePid.Tell(arg)
	}
}

func (rs *RoleActor) PayOrderReq(ctx actor.Context) {
	msg := ctx.Message()
	// 下单
	arg := msg.(*pb.PayOrderReq)
	glog.Debugf("PayOrderReq %#v", arg)
	rs.payOrder(arg)
}

// 提现
func (rs *RoleActor) WithDrawReq(ctx actor.Context) {
	msg := ctx.Message()
	// 下单
	arg := msg.(*pb.PayWithDrawReq)
	glog.Debugf("WithDrawReq %#v", arg)
	rs.withDrawOrder(arg)
}

// 修改提现日志的状态
func (rs *RoleActor) WithdrawLogStatus(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WithdrawLogStatus)
	glog.Debugf("WithdrawLogStatus %#v", arg)
	rs.changeWithdrawStatus(arg)
}

// share充值
func (rs *RoleActor) ShareRecharge(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ShareRecharge)
	glog.Debugf("ShareRecharge %#v", arg)
	if b, ok := rs.ShareBelow[arg.Userid]; ok {
		b.Recharge += int64(arg.Money)
		rs.ShareBelow[arg.Userid] = b
		rs.status = true
	}
}

// func (rs *RoleActor) applePay(arg *pb.ApplePayReq) {
// 	/*rsp, record, trade := handler.AppleOrder(arg, rs.User)
// 	if rsp.Error != pb.OK {
// 		rs.Send(rsp)
// 		return
// 	}
// 	//验证
// 	msg1 := new(pb.ApplePay)
// 	msg1.Trade = trade
// 	timeout := 3 * time.Second
// 	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
// 	if err1 != nil {
// 		glog.Errorf("ApplePay err: %v", err1)
// 		rsp.Error = pb.AppleOrderFail
// 		rs.Send(rsp)
// 		return
// 	}
// 	if response1, ok := res1.(*pb.ApplePaid); ok {
// 		if !response1.Result {
// 			glog.Error("ApplePay fail")
// 			rsp.Error = pb.AppleOrderFail
// 			rs.Send(rsp)
// 			return
// 		}
// 	} else {
// 		glog.Error("ApplePay fail")
// 		rsp.Error = pb.AppleOrderFail
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.sendGoods(int64(record.Diamond), 0, record.Money, record.First)
// 	rs.Send(rsp)*/

// 	// TODO 充值没接先直接跳过
// 	// rsp := new(pb.ApplePayRsp)
// 	// rsp.Id = arg.Id
// 	// d := config.GetShop(utils.String(arg.Id))
// 	// rs.sendGoods(int64(d.Number), int64(d.Give), uint32(d.Price*100), 0, "")
// 	// rs.Send(rsp)
// }

// func (rs *RoleActor) wxPay(arg *pb.WxpayOrderReq) {
// 	var ip string = rs.User.LoginIP
// 	rsp, trade := handler.WxOrder(arg, rs.User, ip)
// 	if rsp.Error != pb.OK {
// 		rs.Send(rsp)
// 		return
// 	}
// 	//验证
// 	msg1 := new(pb.ApplePay)
// 	msg1.Trade = trade
// 	timeout := 3 * time.Second
// 	res1, err1 := rs.rolePid.RequestFuture(msg1, timeout).Result()
// 	if err1 != nil {
// 		glog.Errorf("wxPay err: %v", err1)
// 		rsp.Error = pb.PayOrderFail
// 		rs.Send(rsp)
// 		return
// 	}
// 	if response1, ok := res1.(*pb.ApplePaid); ok {
// 		if !response1.Result {
// 			glog.Error("wxPay fail")
// 			rsp.Error = pb.PayOrderFail
// 			rs.Send(rsp)
// 			return
// 		}
// 	} else {
// 		glog.Error("wxPay fail")
// 		rsp.Error = pb.PayOrderFail
// 		rs.Send(rsp)
// 		return
// 	}
// 	//下单成功
// 	rs.Send(rsp)
// 	//主动查询发货
// 	go rs.wxPayQuery(rsp.Orderid)
// }

// 主动查询发货
// func (rs *RoleActor) wxPayQuery(orderid string) {
// 	//查询
// 	result := handler.ActWxpayQuery(orderid) //查询
// 	if result == "" {
// 		return
// 	}
// 	if rs.rolePid == nil {
// 		return
// 	}
// 	//发货
// 	msg2 := new(pb.WxpayCallback)
// 	msg2.Result = result
// 	rs.rolePid.Tell(msg2)
// }

// 发货
func (rs *RoleActor) sendGoodByType(ctype uint32, number int64, money uint32, logType int32, desc string) {
	switch ctype {
	case data.DIAMOND:
		rs.sendGood(0, 0, number, 0, money, 0, logType, desc)
		// case data.COIN:
		// 	rs.sendGood(0, number, money, 0, logType, desc)
	}
}

// 发货
// func (rs *RoleActor) sendGoods(diamond, coin int64, money uint32, first int, desc string) {
// 	rs.sendGood(diamond, coin, money, first, int32(pb.LOG_TYPE4), desc)
// }

// 发货
func (rs *RoleActor) sendGood(diamond, coin, give, out int64, money uint32, first int, logType int32, desc string) {
	//消息
	rs.addCurrency(diamond, coin, 0, 0, give, out, 0, logType, desc, "")

	//消息
	if money > 0 {
		ntf := &pb.MoneyChangeNtf{Money: int64(rs.User.Money)}
		rs.Send(ntf)

		rs.User.AddMoney(money) // 同步money到数据库
		rs.User.AddChargeTimes()
		rs.pbGift()
	}

	if give != 0 {
		// VB记录日志
		vb := data.VipBankLog{Date: utils.BsonNow().Unix(), LogType: logType, Amount: give}
		rs.VBankLog = append(rs.VBankLog, vb)

		size := len(rs.VBankLog)
		if size > 100 {
			rs.VBankLog = rs.VBankLog[size-100:]
		}
		// 弹窗
		bean := &pb.VipBankLog{
			Date:   vb.Date,
			Ltype:  handler.GetBonusChangeType(vb.LogType, rs.VBBank),
			Amount: int32(vb.Amount),
			Vb:     rs.VBBank,
		}
		ntf := &pb.VipBankLogAddNtf{Log: bean}
		rs.Send(ntf)
	}
}

// 下单
func (rs *RoleActor) payOrder(arg *pb.PayOrderReq) {
	rsp := new(pb.PayOrderRsp)
	// 手机是否在黑名单
	if handler.RechargeInterceptor(rs.Phone2) {
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)
		return
	}
	now := time.Now().In(location)
	// 判断刷子阻拦弹窗
	payNoBrushTime := table.GetTables().PaymentTable.Get().PayNoBrushTime[rs.RegistArea]
	payBrushLimit := table.GetTables().PaymentTable.Get().PayBrushLimit[rs.RegistArea]
	if payNoBrushTime > 0 && payBrushLimit > 0 {
		pay_stats := make(map[string]any)
		start := now.Add(-(time.Duration(payNoBrushTime) * time.Minute))
		// 弹了之后,下次拉单不阻止
		if start.Unix() < rs.User.LastPayOrderBrushLimitTime {
			start = time.Unix(rs.User.LastPayOrderBrushLimitTime, 0).In(location)
		}

		args := []any{rs.Userid, start}
		sql := `select SUM(case when order_status = 1 then 1 else 0 end) pay_tradings, 
			SUM(case when order_status = 4 then 1 else 0 end) pay_success
			from game.col_trade_record final where userid = ? and ctime > ?`
		if err := ck.Select(&pay_stats, sql, args...); err != nil {
			glog.Errorf("select user pay no brush error: %s, %v", rs.Userid, err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
		pay_tradings := utils.ToInt64(pay_stats["pay_tradings"])
		pay_success := utils.ToInt64(pay_stats["pay_success"])
		// 刷单限制弹窗
		if pay_success == 0 && pay_tradings >= int64(payBrushLimit) {
			rs.User.LastPayOrderBrushLimitTime = now.Unix()
			payHelpReward := table.GetTables().PaymentTable.Get().PayHelpReward[rs.RegistArea]
			glog.Infof("user pay brush limit: %s, brush=%d, limit=%d", rs.Userid, pay_tradings, payBrushLimit)
			rsp.Error = pb.PayOrderBrushLimit
			rsp.UnpaidOrders = int32(pay_tradings)
			rsp.SupportReword = int64(payHelpReward)
			rs.Send(rsp)
			return
		}
	}

	// 判断新玩家拉单限制
	newerPayTimesLimit := table.GetTables().PaymentTable.Get().NewerPayTimesLimit[rs.RegistArea]
	newerPaysLimit := table.GetTables().PaymentTable.Get().NewerPaysLimit[rs.RegistArea]
	if newerPayTimesLimit > 0 && newerPaysLimit > 0 {
		var pay_success_times int64
		sql := `select count(*) pay_success_times from game.col_trade_record final where userid = ? and order_status = 4`
		if err := ck.Select(&pay_success_times, sql, rs.Userid); err != nil {
			glog.Errorf("select user pay success times error: %s, %v", rs.Userid, err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
		if pay_success_times <= int64(newerPayTimesLimit) {
			if arg.Amount > newerPaysLimit {
				glog.Infof("newer user pay success times limit: %s, pay_success_times=%d, order_amount=%d", rs.Userid, pay_success_times, arg.Amount)
				rsp.Error = pb.PayOrderFail
				rs.Send(rsp)
				return
			}
		}
	}

	// 判断能不能下订单
	var payMent config.IPay
	switch arg.Rtype {
	case pb.GAMERECHARGE:
		// 去游戏获取局内充值金额
		result, err := rs.gamePid.RequestFuture(&pb.GameRechargeAmount{Userid: rs.Userid}, 5*time.Second).Result()
		if err != nil {
			glog.Info("get amount fail,err:", err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
		if r, ok := result.(*pb.GameRechargeAmounted); ok {
			payMent = &data.DeskRecharge{Price: r.Amount, Give: r.GiveAmount, Desc: arg.Body}
		} else {
			glog.Info("get amount fail,err:", err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
	case pb.BREAKING:
		// 破产礼包
		payMent = handler.GetBreakingPayMent(rs.User)
	case pb.CUSTOM:
		// 自定义充值(支付金额限制改为通道配置判断)
		payMent = &data.CustomShop{Amount: uint32(arg.Amount)}
	case pb.COUPON:
		// 优惠券
		result, err := rs.rolePid.RequestFuture(&pb.GetCoupon{Userid: rs.Userid, CouponId: arg.Id}, 3*time.Second).Result()
		if err != nil {
			glog.Error("coupon pay fail,err:", err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
		if r, ok := result.(*pb.GetCouponed); ok {
			payMent = &data.CouponPay{Amount: int64(arg.Amount), MinRecharge: r.MinAmount, CouponAmount: r.CouponAmount, Id: r.CouponId}
		} else {
			glog.Error("coupon pay fail,err:", err)
			rsp.Error = pb.PayOrderFail
			rs.Send(rsp)
			return
		}
	default:
		payMent = handler.GetPayMent(arg.Id, arg.Rtype)
	}
	if payMent == nil || !payMent.CanOrder(rs.User, arg.Id, int32(arg.Rtype)) {
		glog.Errorf("user %s can't create order", rs.Userid)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)
		return
	}
	// 创建订单
	order, err := handler.CreateOrder(rs.User, payMent, rs.gamePid != nil, rs.gameId, arg.PayWay, arg.PayOption, arg.PayApp)
	if err != nil {
		glog.Errorf("user %s create order fail", rs.Userid)
		glog.Error(err)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)

		rs.RechargeReport(pb.CanNotCreateOrder, err.Error(), arg.LogId)
		return
	}
	order.InGtype = rs.gtype
	// 添加订单记录
	order.ReportId = arg.LogId
	body, err1 := jsoniter.Marshal(order)
	if err1 != nil {
		glog.Errorf("user %s serialize order fail", rs.Userid)
		glog.Error(err)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)
		return
	}
	// 订单是否保存成功
	if !rs.saveOrder(1, 1, body, order.OrderID) {
		glog.Errorf("user %s save order fail", rs.Userid)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)

		rs.RechargeReport(pb.CanNotCreateOrder, "save order fail", arg.LogId)
		return
	}

	rs.RechargeReport(pb.CreateOrderSuccess, fmt.Sprintf("create order success id:%s", order.OrderID), arg.LogId)

	if rs.ToggleChannel >= 5 {
		// 下单5次不充值就换渠道
		rs.ToggleChannel = 0
		rs.CommonChannel = 0
	}
	// 提交订单
	url, err := submit(rs.User, order)
	glog.Infof("下单耗时: user=%s, channel=%d %dms", rs.Userid, order.ChannelId, utils.BsonNow().UnixMilli()-now.UnixMilli())
	if err != nil {
		// 下单失败
		glog.Errorf("submit order fail, user=%s, %v", rs.Userid, err)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)

		// 失败了，更新订单
		order.OrderStatus = data.TradeOrderFail
		order.RefuseReason = err.Error()
		body, err1 = jsoniter.Marshal(order)
		if err1 == nil {
			rs.saveOrder(1, 2, body, order.OrderID)
		}

		return
	} else {
		rsp.Payreq = url
		rsp.Id = arg.Id
		rsp.Orderid = fmt.Sprint(order.OrderID)
		rs.Send(rsp)

		rs.RechargeReport(pb.SubmitOrderSuccess, fmt.Sprintf("orderid:%s,amount:%d,url:%s", order.OrderID, order.Amount, url), arg.LogId)
		glog.Infof("PayOrderRsp: %v", rsp)
		glog.Infof("user %s submit order success, url:%s", rs.Userid, url)
	}

	// 更新订单
	body, err1 = jsoniter.Marshal(order)
	if err1 != nil {
		glog.Errorf("user %s serialize order fail", rs.Userid)
		glog.Error(err)
		rsp.Error = pb.PayOrderFail
		rs.Send(rsp)
		return
	}
	rs.saveOrder(1, 2, body, order.OrderID)

	rs.ToggleChannel++
	rs.status = true
	// test
	// 直接回调
	if env == "dev" && order.PayWay != 1 {
		// up := new(pb.PayOrderUpdate)
		// up.OutId = "111111"
		// up.OrderId = order.OrderID
		// up.State = data.TradeSuccess
		// up.Otype = 1
		// up.OutStatus = "00"
		// rs.rolePid.Tell(up)
	}
}

// 提交订单
func submit(user *data.User, order *data.PayOrder) (string, error) {
	req := data.PayRequest{
		OrderID:    order.OrderID,
		Userid:     order.Userid,
		Amount:     order.Amount,
		RegistIp:   order.RegistIp,
		BusinessID: order.ReportId,
		Channel:    user.CommonChannel,
		PayWay:     order.PayWay,
		PayOption:  order.PayOption,
		PayApp:     order.PayApp,
	}
	body, err := jsoniter.Marshal(req)
	if err != nil {
		return "", err
	}
	rsp, err1 := handler.SubmitOrder(payurl, body, true)
	if err1 != nil {
		return "", err1
	}
	var res *data.PayResponse
	var ok bool
	if res, ok = rsp.(*data.PayResponse); !ok {
		return "", errors.New("recharge order fail, can't convert PayResponse")
	}
	order.ChannelId = res.ChannelID
	order.OutTradeStatus = utils.String(res.Code)
	order.OutTradeNo = res.OrderID
	if res.Code != 200 {
		return "", fmt.Errorf("recharge order fail, code:%d,msg:%s", res.Code, res.Msg)
	}
	return res.PaymentUrl, nil

	// channelId := order.ChannelId //渠道id
	// // 根据渠道id选择支付方式
	// payment := handler.GetPayPartenr(channelId)
	// if payment == nil {
	// 	return "", fmt.Errorf("no found payPartenr, channelId:%d", channelId)
	// }
	// // 提交订单
	// body, err := payment.Submit(order)
	// if err != nil {
	// 	return "", err
	// }
	// return payment.SubmitResponse(body)
	// result := new(mlpay.MlOrderRespond)
	// jsoniter.Unmarshal(body, result)
	// if result.Code != "0000" { // 没成功TODO
	// 	return "", fmt.Errorf("code:%s, message:%s", result.Code, result.Message)
	// }

	// return result.Data, nil
}

// 充值之后活动变化
func (rs *RoleActor) paySuccess(arg *pb.PayGoods, ctx actor.Context) error {
	rs.TpUserTodayChargeNum++
	rs.TpUserTodayChargeAmount += int64(arg.Money)

	rs.ToggleChannel = 0
	rs.CommonChannel = arg.ChannelId // 设置为常用渠道
	rsp := handler.PaySuccess(rs.User, arg.ShopId, arg.ReportId, arg.ShopType, arg.Money, false)
	rs.Send(rsp)

	// 充值任务事件
	ntf := event.Event(rs.User, event.RECHARGE_TASK, &event.RechargeTaskEvent{Amount: arg.Money})
	if ntf != nil {
		rs.Send(ntf)
	}

	// 商城罐子
	ntf = event.Event(rs.User, event.RECHARGE_JAR, &event.RechargeJarEvent{ShopType: int(arg.ShopType), ShopId: arg.ShopId, Give: arg.Coin})
	if ntf != nil {
		rs.Send(ntf)
	}

	// vip
	// ntf = event.Event(rs.User, event.VIP, &event.VIPEvent{Amount: int64(arg.Money)})
	// if ntf != nil {
	// 	rs.Send(ntf)
	// }

	// 小米手机活动
	// if arg.Money >= 50000 {
	// 	rs.rolePid.Tell(&pb.LuckyDrawNumber{
	// 		Userid:      rs.Userid,
	// 		ChargeMoney: int64(arg.Money),
	// 	})
	// }

	if rs.RegistArea == 1 && (rs.State == 1 || rs.State == 3) {
		rs.BGiveCash += arg.Diamond
	}

	rs.checkUserState()
	// 状态检测事件
	state := event.Event(rs.User, event.CHECK_STATE, &event.CheckStateEvent{})
	if state != nil {
		if rs.gamePid != nil {
			rs.gamePid.Request(state, ctx.Self())
		}
		if rs.cpPid != nil {
			rs.cpPid.Request(state, ctx.Self())
		}
	}

	// 充值完成事件
	if config.SettingIsOpen(4, data.MAILREPLY) {
		ntf = event.Event(rs.User, event.RECHARGE_MAIL, &event.RechargeMailEvent{Amount: arg.Money})
		if ntf != nil {
			rs.Send(ntf)
		}
	}
	// 分享
	if rs.ShareSuperior != "" {
		share := &pb.ShareRecharge{Money: int32(arg.Money), Userid: rs.Userid, Superior: rs.ShareSuperior}
		rs.rolePid.Tell(share)
	}

	if arg.ShopType == data.SHOP {
		bonus := handler.GetBonusByShop(rs.RegistArea, arg.ShopId)
		// 日志
		rs.BonusLog(rs.Userid, bonus)
	}

	// 回填utr奖励
	if arg.LastUtr != "" {
		msg := &pb.PayUTRBackReword{
			Userid:     rs.Userid,
			RegistArea: int32(rs.RegistArea),
			Amount:     int64(arg.Money),
			OrderId:    arg.Orderid,
			LastUtr:    arg.LastUtr,
		}
		rs.rolePid.Tell(msg)
	}

	if rs.ShareSuperior != "" && !rs.ShareAgent.ShareSuper {
		glog.Infof("user recharge, share superior: %s, userid: %s, amount: %d", rs.ShareSuperior, rs.Userid, arg.Money)
		rs.rolePid.Tell(&pb.ShareAgentPay{
			Userid:  rs.Userid,
			SuperId: rs.ShareSuperior,
			Amount:  int64(arg.Money),
			OrderId: arg.Orderid,
		})
	}

	rs.status = true
	return nil
}

// 保存订单到db  otype:(1:充值 2:提现) stype:(1:新增 2:更新)
func (rs *RoleActor) saveOrder(otype, stype int32, body []byte, order string) bool {
	msg := new(pb.PayOrder)
	msg.Order = body
	msg.Stype = stype
	msg.Otype = otype
	reslut, err2 := rs.rolePid.RequestFuture(msg, time.Second*5).Result()
	if err2 != nil {
		glog.Errorf("user %s save order fail, orderId: %s, err:%v", rs.Userid, order, err2)
		return false
	}
	s := reslut.(*pb.PayedOrder)
	if !s.Result {
		glog.Errorf("user %s save order fail, orderId: %s,", rs.Userid, order)
		return false
	}
	return true
}

// ======================================================提现====================================================== //
func (rs *RoleActor) withDrawOrder(arg *pb.PayWithDrawReq) {
	user := rs.User
	rsp := new(pb.PayWithDrawRsp)
	// 绑没绑定手机
	if user.Phone == "" {
		rsp.Error = pb.NoBindPhone
		rs.Send(rsp)
		return
	}
	// 能不能下单
	if ok, code, param := handler.CanWithDraw(user, arg.Amount); !ok {
		glog.Errorf("user %s can't create withdraworder", rs.Userid)
		rsp.Error = code
		rsp.Param = param
		rs.Send(rsp)
		return
	}

	// 今日提现次数金额限制二次确认
	now := time.Now().In(location)
	stime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	etime := stime.AddDate(0, 0, 1)
	var withdraw_data = make(map[string]any)
	err := ck.Select(&withdraw_data, `
		select count(*) withdraw_times, SUM(amount) withdraw_amounts
		from game.col_withdraw_record final 
		where ctime >= ? and ctime < ? and userid = ? and order_status not in (3, 12)
	`, stime, etime, user.Userid)
	if err != nil {
		glog.Errorf("user %s withdraw order select error: ", err)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	} else {
		withdraw_times := utils.ToInt64(withdraw_data["withdraw_times"])
		withdraw_amounts := utils.ToInt64(withdraw_data["withdraw_amounts"])

		var errCode pb.ErrCode
		vip := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
		if withdraw_times >= int64(vip.WithdrawTimes) && vip.WithdrawTimes != -1 {
			// 提现次数不足
			errCode = pb.WithdrawCountUnenough
		}
		if (int64(arg.Amount) > vip.WithdrawAmounts ||
			withdraw_amounts >= vip.WithdrawAmounts ||
			int64(arg.Amount)+withdraw_amounts > vip.WithdrawAmounts) &&
			vip.WithdrawAmounts != -1 {
			// 提现金额不足
			errCode = pb.WithdrawAmountUnenough
		}
		if errCode != pb.OK {
			glog.Errorf("user %s can't create withdraworder second check: vip=%d, db=(%d, %d), user=(%d, %d)", rs.Userid, user.Vip.Lv,
				withdraw_times, withdraw_amounts,
				user.Vip.WithdrawCount, user.Vip.WithdrawAmount,
			)
			rsp.Error = errCode
			rs.Send(rsp)
			return
		}
	}

	// 保存银行信息(不可修改的)
	rs.saveBank(arg)
	// 创建订单
	order, err := handler.CreateWithDrawOrder(rs.User, arg)
	if err != nil {
		glog.Errorf("user %s create withdraworder fail", rs.Userid)
		glog.Error(err)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	// 保存订单记录
	body, err1 := jsoniter.Marshal(order)
	if err1 != nil {
		glog.Errorf("user %s create serialize fail, err:%v", rs.Userid, err1)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	if !rs.saveOrder(2, 1, body, order.OrderID) {
		glog.Errorf("user %s save withdraw order fail, err:%v", rs.Userid, err1)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	// 扣钱
	rs.addCurrency(-int64(order.Score+order.Commission), 0, 0, 0, 0, -int64(order.Score), 0, int32(pb.LOG_TYPE13), "提现", order.OrderID)
	// rs.sendGood(-int64(order.Score), 0, 0, -int64(order.Score), 0, 0, int32(pb.LOG_TYPE13), "提现")
	// 增加提现次数
	if user.WithDrawCountMap == nil {
		user.WithDrawCountMap = make(map[int32]int32)
	}
	count := user.WithDrawCountMap[arg.Id]
	user.WithDrawCountMap[arg.Id] = count + 1
	user.Vip.WithdrawCount++
	user.Vip.WithdrawAmount += int(arg.Amount)

	user.TpUserTodayWithdrawNum++
	user.TpUserTodayWithdrawAmount += int64(order.Score)

	rsp.Id = arg.Id
	rsp.Cash = user.Diamond
	rsp.BankName = arg.BlankAccount
	rsp.Ifsc = arg.IFSC
	rsp.BankNumber = arg.BlankNumber
	rsp.PayWay = arg.PayWay
	rsp.Count = count + 1
	rsp.WithdrawCount = int32(user.Vip.WithdrawCount)
	rsp.WithdrawAmount = int32(user.Vip.WithdrawAmount)
	rs.Send(rsp)
	// 添加订单记录
	rs.addWithdrawLog(user, order)
	// 发邮件
	ntf := new(pb.FeedBackLogNtf)
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: rs.Userid,
		Title:    "System Message",
		Content:  handler.BuildWithdrawApply(user.Nickname),
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	user.FeedBackLogMap[chat.Uid] = chat
	rs.status = true
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	rs.Send(ntf)

	// 提现下单跑马灯(关闭)
	// photo, _ := strconv.Atoi(rs.Photo)
	// marquee := &pb.MarQueeWithdrawNtf{
	// 	Userid:   rs.Userid,
	// 	NickName: rs.Nickname,
	// 	Photo:    int32(photo),
	// 	VipLv:    int32(rs.Vip.Lv),
	// 	Amount:   int32(order.Score),
	// }
	// Content: handler.BuildWithdrawApplyMarquee(rs.Nickname, int64(order.Score)),
	// rs.rolePid.Tell(marquee)
}

// 保存银行信息
func (rs *RoleActor) saveBank(arg *pb.PayWithDrawReq) {
	user := rs.User
	user.BackType = arg.PayWay
	if arg.PayWay == 1 {
		user.USDT = arg.BlankNumber
		rs.status = true
		return
	}
	if user.Bank != "" {
		return
	}
	user.Bank = arg.BlankAccount
	// 把空格替换掉
	user.BankAccounts = strings.ReplaceAll(arg.BlankNumber, " ", "")
	user.IFSC = strings.ReplaceAll(arg.IFSC, " ", "")
	rs.status = true
}

// 保存银行信息
func (rs *RoleActor) saveBank2(arg *pb.SaveBankInfoReq) {
	user := rs.User
	user.Bank = arg.BlankAccount
	user.BankAccounts = arg.BlankNumber
	user.IFSC = arg.IFSC
	user.BankAccountHolder = arg.BlankAccountHolder
	rs.status = true
}

// 保存usdt信息
func (rs *RoleActor) saveBank3(arg *pb.SaveBankInfoReq) {
	user := rs.User
	user.USDT = arg.BlankNumber
	rs.status = true
}

// 新增提现日志
func (rs *RoleActor) addWithdrawLog(user *data.User, order *data.WithdrawOrder) {
	if user.WithdrawLogMap == nil {
		user.WithdrawLogMap = make(map[string]*data.WithdrawLog)
	}
	log := &data.WithdrawLog{
		Id:     order.OrderID,
		Ctime:  utils.LocalTime().Unix(),
		Amount: order.Amount,
		Status: data.Withdrawing,
		PayWay: order.PayWay,
	}
	user.WithdrawLogMap[order.OrderID] = log
	user.CashOut += int32(order.Amount)
	rs.status = true
	// 通知客户端
	rs.notifyWithdrawLog(log)
}

// 改变日志状态
func (rs *RoleActor) changeWithdrawStatus(arg *pb.WithdrawLogStatus) {
	user := rs.User
	user.CashOut += arg.Amount
	if log, ok := user.WithdrawLogMap[arg.OrderId]; ok {
		log.Status = int(arg.Status)
		if int(arg.Status) == data.WithdrawFreeze {
			log.Status = data.WithdrawSuccess
		}

		// if int(arg.Status) == data.WithdrawInspect {
		// 	log.Status = data.Withdrawing
		// }
		// 更新
		rs.notifyWithdrawLog(log)
		// 邮件
		if config.SettingIsOpen(4, data.MAILREPLY) {
			ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: int(arg.Status)})
			if ntf != nil {
				rs.Send(ntf)
			}
		}

		// 银行信息错误的邮件
		// if int(arg.Status) == data.WithdrawBackBankErr {
		// 	ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: data.WithdrawBackBankErr})
		// 	if ntf != nil {
		// 		rs.Send(ntf)
		// 	}
		// }
		// 回退需要把提现次数加回来
		if arg.Status == data.WithdrawBack || arg.Status == data.WithdrawRefunded {
			ctime := utils.Stamp2Time(log.Ctime)

			if utils.Time2DayDate(ctime.In(location)) == utils.Time2DayDate(time.Now().In(location)) {
				// 同一天才退次数
				if user.Vip.WithdrawCount > 0 {
					user.Vip.WithdrawCount--
					user.Vip.WithdrawAmount += int(arg.Amount)
				}
			}
		}
	}

	if arg.Status == data.WithdrawSuccess { // 巴基斯坦
		rs.SavePKWithdrawInfo(arg.GetBankName())
	}
	rs.status = true
}

func (rs *RoleActor) notifyWithdrawLog(log *data.WithdrawLog) {
	// 通知客户端
	rsp := new(pb.AddWithdrawLogNtf)
	rsp.Logs = &pb.WithdrawLog{
		Id:     log.Id,
		Time:   log.Ctime,
		Amount: log.Amount,
		Status: pb.WithdrawLog_State(log.Status - 1),
		PayWay: log.PayWay,
	}
	rs.Send(rsp)
}

// 充值埋点
func (rs *RoleActor) RechargeReport(code pb.RechargeReportReq_Code, content, uid string) {
	log := &pb.LogRechargeReport{
		Uid:     uid,
		Code:    int32(code),
		Content: content,
	}
	myactor.Logger().Tell(log)
}

// 充值记录查询
func (rs *RoleActor) PaymentRecordReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PaymentRecordReq)
	rsp := &pb.PaymentRecordRsp{}
	glog.Debugf("PaymentRecordReq %#v", arg)

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := today
	switch arg.Range {
	case 1:
		start = today
	case 2:
		start = today.AddDate(0, 0, -7)
	case 3:
		start = today.AddDate(0, 0, -30)
	}
	rsp.DateRange = fmt.Sprintf("%s to %s", start.Format(utils.FORMAT_DATE), now.Format(utils.FORMAT_DATE))

	if arg.PageSize <= 0 || arg.PageSize > 100 {
		arg.PageSize = 100
	}

	// 查询充值成功总和
	sql_pays := `
		select SUM(amount) pays
		from game.col_trade_record final
		where userid = ? and ctime >= ? and order_status = 4
	`
	if err := ck.Select(&rsp.PaySum, sql_pays, rs.Userid, start); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	// 查询记录详情
	sql := `
		select order_id, pay_way, ctime, order_status, amount, score, shop_type, other_present, last_utr, give_cash
		from game.col_trade_record final
		where userid = ? and ctime >= ? %s order by ctime desc limit ?
	`
	args := []any{rs.Userid, start}
	if arg.PrevOrderId != "" {
		sql = fmt.Sprintf(sql, `and ctime < (select ctime from game.col_trade_record final where order_id = ?)`)
		args = append(args, arg.PrevOrderId)
	} else {
		sql = fmt.Sprintf(sql, "")
	}
	args = append(args, arg.PageSize)

	var results []map[string]any
	if err := ck.Select(&results, sql, args...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	// utr
	orderids := []string{}
	utrs := []data.Utr{}
	utrMap := make(map[string]data.Utr)
	for _, r := range results {
		orderids = append(orderids, fmt.Sprint(r["order_id"]))
	}
	data.ListByQ(data.Utrs, bson.M{"orderid": bson.M{"$in": orderids}}, &utrs)
	for _, u := range utrs {
		utrMap[u.Orderid] = u
	}

	payingTimeLimit := table.GetTables().PaymentTable.Get().PayingTimeLimit
	payCustomTime := table.GetTables().PaymentTable.Get().PayCustomTime[rs.RegistArea]
	for _, r := range results {
		order_id := fmt.Sprint(r["order_id"])
		pay_way := utils.ToInt64(r["pay_way"])
		ctime := (r["ctime"].(time.Time)).In(location)
		order_status := utils.ToInt64(r["order_status"])
		amount := utils.ToInt64(r["amount"])
		give_cash := utils.ToInt64(r["give_cash"])
		shop_type := utils.ToInt64(r["shop_type"])
		other_present := utils.ToInt64(r["other_present"])
		last_utr := fmt.Sprint(r["last_utr"])
		record := &pb.PaymentRecord{
			OrderId:    order_id,
			PayWay:     int32(pay_way),
			PayTime:    ctime.Format(utils.FORMAT),
			Amount:     amount,
			GiveNumber: int32(other_present),
			UserUTR:    last_utr,
			UtrStatus:  3, // 默认未上传
			ShopType:   int32(shop_type),
		}
		if shop_type == data.COUPON {
			record.GiveCash = int32(give_cash)
		}
		if u, ok := utrMap[order_id]; ok {
			record.UtrStatus = int32(u.Status)
		}
		if order_status == data.TradeSuccess {
			// success
			record.Status = 2
		} else if order_status == data.Tradeing {
			// pending
			if now.Before(ctime.Add(time.Duration(payingTimeLimit) * time.Minute)) {
				record.Status = 1
			} else {
				record.Status = 3
			}
		} else {
			// timeout/failed
			record.Status = 3
		}
		if record.Status == 2 || record.Status == 3 {
			if now.Before(ctime.Add(time.Duration(payCustomTime) * time.Hour)) {
				record.Custom = true
			}
		}
		rsp.Records = append(rsp.Records, record)
	}

	rs.Send(rsp)
}

// 提现记录查询
func (rs *RoleActor) WithdrawRecordReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordReq)
	rsp := &pb.WithdrawRecordRsp{}
	glog.Debugf("WithdrawRecordReq %#v", arg)

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := today
	switch arg.Range {
	case 1:
		start = today
	case 2:
		start = today.AddDate(0, 0, -7)
	case 3:
		start = today.AddDate(0, 0, -30)
	}
	rsp.DateRange = fmt.Sprintf("%s to %s", start.Format(utils.FORMAT_DATE), now.Format(utils.FORMAT_DATE))

	if arg.PageSize <= 0 || arg.PageSize > 100 {
		arg.PageSize = 100
	}

	// 查询充值成功总和
	sql_withdraws := `
		select SUM(amount) withdraws
		from game.col_withdraw_record final
		where userid = ? and ctime >= ? and order_status = 2
	`
	if err := ck.Select(&rsp.PaySum, sql_withdraws, rs.Userid, start); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	sql := `
		select order_id, ctime, pay_time, order_status, amount, score, commission, 
			overtime_pay_tacks, keep_wait_time
		from game.col_withdraw_record final 
		where userid = ? and ctime >= ? %s order by ctime desc limit ?
	`
	args := []any{rs.Userid, start}
	if arg.PrevOrderId != "" {
		sql = fmt.Sprintf(sql, `and ctime < (select ctime from game.col_withdraw_record final where order_id = ?)`)
		args = append(args, arg.PrevOrderId)
	} else {
		sql = fmt.Sprintf(sql, "")
	}
	args = append(args, arg.PageSize)

	var results []map[string]any
	if err := ck.Select(&results, sql, args...); err != nil {
		rsp.Error = pb.Failed
		rs.Send(rsp)
		return
	}

	// 提现保障时效（小时）
	withdrawSafeguardTime := table.GetTables().PaymentTable.Get().WithdrawSafeguardTime[rs.RegistArea]
	// 超时赔付金
	withdrawTimeoutPayment := table.GetTables().PaymentTable.Get().WithdrawTimeoutPayment[rs.RegistArea]
	withdrawOvertimeBeginDateS := table.GetTables().PaymentTable.Get().WithdrawOvertimeBeginDate
	withdrawOvertimeBeginDate, overtimeErr := time.ParseInLocation(utils.FORMAT_DATE, withdrawOvertimeBeginDateS, location)
	if overtimeErr != nil {
		glog.Errorf("withdrawOvertimeBeginDate parse error: %s, %v", withdrawOvertimeBeginDate, overtimeErr)
	}
	// 超时赔偿类型和金额
	var overtimePayType, overtimePay int32
	for t, value := range withdrawTimeoutPayment.Value {
		if value > 0 {
			overtimePay = value
			overtimePayType = int32(t + 1)
			break
		}
	}

	for _, r := range results {
		order_id := fmt.Sprint(r["order_id"])
		ctime := (r["ctime"].(time.Time)).In(location)
		pay_time := (r["pay_time"].(time.Time)).In(location) // 订单完结时间
		order_status := utils.ToInt64(r["order_status"])
		amount := utils.ToInt64(r["amount"])
		overtime_pay_tacks := utils.ToInt64(r["overtime_pay_tacks"])
		keep_wait_time := utils.ToInt64(r["keep_wait_time"])

		record := &pb.WithdrawRecord{
			OrderId:      order_id,
			WithdrawTime: ctime.Format(utils.FORMAT),
			Score:        amount,
		}
		if order_status == data.WithdrawSuccess {
			record.Status = 3
		} else if order_status == data.OrderSuccess ||
			order_status == data.WithdrawUrgentTransfer ||
			order_status == data.WithdrawTimeoutTransfer {
			record.Status = 2
		} else if order_status == data.Withdrawing ||
			order_status == data.WaitDelivery {
			record.Status = 1
		} else if order_status == data.WithdrawRefunded ||
			order_status == data.WithdrawBack {
			record.Status = 5
		} else {
			record.Status = 4
		}

		// 超时赔付计算
		if overtimePay > 0 &&
			withdrawSafeguardTime > 0 &&
			overtimeErr == nil &&
			ctime.After(withdrawOvertimeBeginDate) {
			finishTime := now // 完结时间
			if !utils.SliceIn(record.Status, 1, 2) {
				finishTime = pay_time
			}
			passHours := finishTime.Sub(ctime).Hours()
			if passHours > 0 {
				// 计算赔偿次数
				overtimePayTimes := int32(math.Floor(passHours / float64(withdrawSafeguardTime)))
				// 赔偿金
				overtimePays := overtimePay * overtimePayTimes

				record.OvertimePayType = overtimePayType
				record.OvertimePay = int64(overtimePays) - overtime_pay_tacks
				if record.OvertimePay < 0 {
					record.OvertimePay = 0
				}

				// 赔付倒计时中: 1.Pending 2.BankProcessing
				if utils.SliceIn(record.Status, 1, 2) {
					record.OvertimePayON = true
					// 下次赔偿金
					record.OvertimeWillPay = record.OvertimePay + int64(overtimePay)
					nextOvertimePayTime := ctime
					for {
						nextOvertimePayTime = nextOvertimePayTime.Add(time.Duration(withdrawSafeguardTime) * time.Hour)
						if nextOvertimePayTime.After(now) {
							break
						}
					}
					record.OvertimePayTime = nextOvertimePayTime.Unix()

				} else {
					// 有未领取的赔偿金
					if record.OvertimePay > 0 {
						record.OvertimePayON = true
					}
				}
			}
		}

		// 是否可退款
		if order_status == data.Withdrawing && keep_wait_time == 0 {
			refundTime := table.GetTables().PaymentTable.Get().RefundTime[rs.RegistArea]
			if now.After(ctime.Add(time.Duration(refundTime) * time.Hour)) {
				refundCompensateRate := table.GetTables().PaymentTable.Get().RefundCompensateRate[rs.RegistArea]
				refundCompensateAmount := table.GetTables().PaymentTable.Get().RefundCompensateAmount[rs.RegistArea]
				refundCompensateType := table.GetTables().PaymentTable.Get().RefundCompensateType[rs.RegistArea]

				record.Refund = true
				record.RefundCompensate = int64(math.Floor(float64(amount)*(float64(refundCompensateRate)/10000))) + int64(refundCompensateAmount)
				record.RefundCompensateType = refundCompensateType
			}
		}

		// 客服
		if !utils.SliceIn(record.Status, 1, 2) && !record.OvertimePayON {
			// 客服栏超时时间
			withdrawCustomTime := table.GetTables().PaymentTable.Get().WithdrawCustomTime[rs.RegistArea]
			if now.Before(pay_time.Add(time.Duration(withdrawCustomTime) * time.Hour)) {
				record.Custom = true
			}
		}

		rsp.Records = append(rsp.Records, record)
	}
	rs.Send(rsp)
}

// 充值记录utr回填请求
func (rs *RoleActor) PaymentRecordUTRReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PaymentRecordUTRReq)
	glog.Debugf("PaymentRecordUTRReq %#v", arg)

	arg.Userid = rs.Userid
	rs.rolePid.Request(arg, ctx.Self())
}

// 提现记录超时退款请求
func (rs *RoleActor) WithdrawRecordRefundReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordRefundReq)
	glog.Debugf("WithdrawRecordRefundReq %#v", arg)

	arg.Userid = rs.Userid
	arg.RegistArea = int32(rs.RegistArea)
	rs.rolePid.Request(arg, ctx.Self())
}

// 提现记录超时继续等待请求
func (rs *RoleActor) WithdrawRecordKeepWaitingReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordKeepWaitingReq)
	glog.Debugf("WithdrawRecordKeepWaitingReq %#v", arg)

	arg.Userid = rs.Userid
	arg.RegistArea = int32(rs.RegistArea)
	rs.rolePid.Request(arg, ctx.Self())
}

// 提现记录超时继续等待请求
func (rs *RoleActor) WithdrawRecordOvertimePayTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordOvertimePayTackReq)
	glog.Debugf("WithdrawRecordOvertimePayTackReq %#v", arg)

	arg.Userid = rs.Userid
	arg.RegistArea = int32(rs.RegistArea)
	rs.rolePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) HasFailUtr(ctx actor.Context) bool {
	result, err := rs.rolePid.RequestFuture(&pb.HasFailUtr{}, time.Second).Result()
	if err != nil {
		return false
	}
	if f, ok := result.(pb.HasFailUtred); ok {
		return f.Has
	}
	return false
}

func (rs *RoleActor) WithdrawBankInfoReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WithdrawBankInfoReq)
	glog.Debugf("WithdrawBankInfoReq %#v", arg)
	rs.WithdrawBankInfo(arg)
}
