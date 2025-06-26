package dbms

import (
	"context"
	"fmt"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
)

// 用户下单 (1:充值  2:提现)
func (a *RoleActor) PayOrder(ctx actor.Context) {
	msg := ctx.Message()
	// 充值下单
	arg := msg.(*pb.PayOrder)
	glog.Debugf("PayOrder %#v", arg)
	var rsp *pb.PayedOrder
	if arg.Otype == 1 {
		if arg.Stype == 1 {
			rsp = a.savePayOrder(arg.Order)
		} else if arg.Stype == 2 {
			rsp = a.updatePayOrder(arg.Order)
		}
	} else {
		if arg.Stype == 1 {
			rsp = a.saveWithDrawOrder(arg.Order)
		} else if arg.Stype == 2 {
			rsp = a.updateWithDrawOrder(arg.Order)
		}
	}
	ctx.Respond(rsp)
}

// 下单回调(1:充值  2:提现)
func (a *RoleActor) PayOrderUpdate(ctx actor.Context) {
	msg := ctx.Message()
	// 订单更新
	arg := msg.(*pb.PayOrderUpdate)
	glog.Debugf("PayOrderUpdate %#v", arg)
	switch arg.Otype {
	case 1:
		a.payOrderUpdate(arg, ctx)
	case 2:
		a.withdrawOrderUpdate(arg, ctx)
	default:
		glog.Errorf("otype fail, %d", arg.Otype)
	}
}

/*==================================充值==================================*/
// 充值下单
func (a *RoleActor) savePayOrder(body []byte) (rsp *pb.PayedOrder) {
	order := new(data.TradeRecord)
	rsp = new(pb.PayedOrder)
	err := jsoniter.Unmarshal(body, order)
	if err != nil {
		rsp.Result = false
		return
	}
	if order.Has() {
		rsp.Result = false
		glog.Errorf("tradeOrder has %#v", order)
		return
	}

	if order.Save() {
		rsp.Result = true
		glog.Infof("tradeOrder success %#v", order)
	}
	return
}

func (a *RoleActor) updatePayOrder(body []byte) (rsp *pb.PayedOrder) {
	order := new(data.TradeRecord)
	rsp = new(pb.PayedOrder)
	err := jsoniter.Unmarshal(body, order)
	if err != nil {
		rsp.Result = false
		return
	}
	if !order.Has() {
		rsp.Result = false
		glog.Errorf("tradeOrder no found %#v", order)
		return
	}
	if order.Update() {
		rsp.Result = true
		glog.Infof("tradeOrder update success %#v", order)
	}
	return
}

// 订单更新
func (a *RoleActor) payOrderUpdate(arg *pb.PayOrderUpdate, ctx actor.Context) {
	order := &data.TradeRecord{
		OrderID: arg.OrderId,
	}
	rsp := new(pb.PayedOrderUpdate)
	order.Get()
	if order.OrderStatus != data.Tradeing {
		// 重复处理的订单
		rsp.Result = true
		ctx.Respond(rsp)
		return
	}

	defer order.Upsert()

	order.OrderStatus = arg.GetState()
	order.PayTime = utils.BsonNow()
	order.OutTradeNo = arg.OutId
	order.RefuseReason = arg.Err
	order.OutTradeStatus = arg.OutStatus
	order.Repair = arg.Repair
	order.RealAmount = order.Amount
	if arg.Amount != 0 {
		// 用户实际支付
		order.RealAmount = uint32(arg.Amount)
	}
	if arg.State == data.TradeSuccess && order.ShopType != data.WEB_TEST_SHOP {
		if user := a.getUserById(order.Userid); user != nil {
			order.FirstPay = user.Money <= 0
		}

		// 发货
		a.tradeHandler(order)

		// 最后一次充值成功时间
		a.lastChargeTime = utils.BsonNow().Unix()

		// 埋点
		log := &pb.LogRechargeReport{
			Uid:     order.ReportId,
			Code:    int32(pb.SendReward),
			Content: fmt.Sprintf("%s recharge success,send reward", arg.OrderId),
		}
		myactor.Logger().Tell(log)
	}
	ctx.Respond(rsp)
}

// 支付手动回调
func (a *RoleActor) payManualCallback(orderId string, rsp *pb.WebResponse) {
	order := &data.TradeRecord{
		OrderID: orderId,
	}
	order.Get()
	if order == nil || order.OrderID == "" {
		rsp.ErrMsg = fmt.Sprintf("invalid order: %v", orderId)
		return
	}
	if order.OrderStatus != data.Tradeing {
		// 重复处理的订单
		rsp.ErrMsg = fmt.Sprintf("multiple processing order: %v", orderId)
		return
	}
	order.OrderStatus = data.TradeSuccess
	order.PayTime = utils.BsonNow()
	order.OutTradeStatus = "manual"
	order.Repair = true

	if order.ShopType == data.WEB_TEST_SHOP {
		// 后台测试订单
		if !order.Upsert() {
			glog.Errorf("web test trade save failed: %#v", order)
		}
		return
	}

	if user := a.getUserById(order.Userid); user != nil {
		order.FirstPay = user.Money <= 0
	}
	// 发货
	a.tradeHandler(order)
}

/*==================================提现==================================*/
// 提现下单
func (a *RoleActor) saveWithDrawOrder(body []byte) (rsp *pb.PayedOrder) {
	order := new(data.WithdrawRecord)
	rsp = new(pb.PayedOrder)
	err := jsoniter.Unmarshal(body, order)
	if err != nil {
		rsp.Result = false
		return
	}
	if order.Has() {
		rsp.Result = false
		glog.Errorf("tradeOrder has %#v", order)
		return
	}
	if order.Save() {
		rsp.Result = true
		glog.Errorf("tradeOrder success %#v", order)
	}
	return
}

func (a *RoleActor) updateWithDrawOrder(body []byte) (rsp *pb.PayedOrder) {
	order := new(data.WithdrawRecord)
	rsp = new(pb.PayedOrder)
	err := jsoniter.Unmarshal(body, order)
	if err != nil {
		rsp.Result = false
		return
	}
	if !order.Has() {
		rsp.Result = false
		glog.Errorf("tradeOrder no found %#v", order)
		return
	}
	if order.Update() {
		rsp.Result = true
		glog.Errorf("tradeOrder update success %#v", order)
	}
	return
}

// 提现审核
func (a *RoleActor) withdrawExamine(rsp *pb.WebResponse, op *data.WithdrawOpreate) {
	order := &data.WithdrawRecord{
		OrderID: op.OrderID,
	}
	if !order.Has() {
		rsp.ErrMsg = "no found order"
		return
	}
	order.Get()
	switch op.Op {
	// 通过,请求第三方
	case data.Pass, data.Delivery:
		if order.OrderStatus != data.Withdrawing &&
			order.OrderStatus != data.WaitDelivery {
			rsp.ErrMsg = "重复审核"
			return
		}
		// 判断是否开启自动派单
		if op.Op == data.Pass && !config.SettingIsOpen(4, data.AUTODELIVERY) {
			// 把订单改为待派单
			order.UpdateOrderStatus(data.WaitDelivery)
			return
		}
		for _, detail := range order.TransferDetail {
			if detail.TType == data.NoTransfer {
				rsp.ErrMsg = "重复审核"
				return
			}
		}
		go a.submitWithdrawOrder(order, false, op.UserName, data.NormalTransfer, data.NoTransfer)
	case data.Transfer:
		// 转单
		if order.OrderStatus == data.OrderSuccess ||
			order.OrderStatus == data.WaitDelivery ||
			order.OrderStatus == data.WithdrawSuccess {
			rsp.ErrMsg = "order status error"
			glog.Errorf("order status error, can't transfer. order %s, user:%s", order.OrderID, order.Userid)
			return
		}
		// 判断补贴金够不够
		if !a.checkTransferCash(order) {
			rsp.ErrMsg = "补贴金不够"
			return
		}
		a.transferWithdrawOrder(order, op.UserName, data.HandTransfer)
	case data.Back:
		// 回退
		if a.withdrawOrderBack(order, data.WithdrawBack) != "" {
			rsp.ErrMsg = "order back fail"
		}
	case data.Freeze:
		// 冻结
		err := a.withdrawOrderFreeze(order)
		if err != "" {
			rsp.ErrMsg = err
		}
	}
	// 提现
}

// 提交提现订单
func (a *RoleActor) submitWithdrawOrder(order *data.WithdrawRecord, transfer bool, name string, reason, ttype int) {
	msg := new(pb.PayOrderUpdate)
	msg.OrderId = order.OrderID
	msg.Otype = 2
	body, err := jsoniter.Marshal(order)
	if err != nil {
		glog.Errorf("[withdraw fail] serialize strut fail. order %s, user:%s", order.OrderID, order.Userid)
		msg.Err = "serialize fail"
		msg.State = data.OrderFail
		return
	}
	// 结构体转换
	outOrder := data.CovertWithdrawOrder(body)
	if outOrder.OrderID == "" {
		glog.Errorf("[withdraw fail] convert strut fail. order %s, user:%s", order.OrderID, order.Userid)
		msg.Err = "convert fail"
		msg.State = data.OrderFail
		return
	}
	// 提交第三方
	// channle, _ := strconv.Atoi(outOrder.OutChannel)
	// payment := handler.GetPayPartenr(uint32(channle))
	resp, err1 := a.withdrawSubmit(outOrder, transfer)
	if err1 != nil {
		glog.Errorf("[withdraw fail] submit fail. order %s, user:%s, err:%s", order.OrderID, order.Userid, err1)
		return
	}
	// succ, code := payment.WithdrawSubmitResponse(resp)

	msg.OutId = resp.OrderID
	msg.OutStatus = fmt.Sprintf("%d", resp.Code)

	// result := new(mlpay.MlOrderRespond)
	// jsoniter.Unmarshal(resp, result)
	// if resp.Code == 200 { // 没成功TODO
	// 	// 发条消息提单成功
	// 	glog.Infof("[withdraw] submit success. order %s, user:%s", order.OrderID, order.Userid)
	// 	msg.State = data.OrderSuccess
	// } else {
	// 	glog.Infof("[withdraw fail] submit fail. order %s, user:%s", order.OrderID, order.Userid)
	// 	msg.State = data.OrderFail
	// }

	glog.Infof("[withdraw] order %s, user:%s, status:%d", order.OrderID, order.Userid, resp.Status)
	msg.State = resp.Status

	channel := strconv.Itoa(int(resp.ChannelID))
	if order.OutChannel != channel {
		order.OutChannel = channel
		order.ETime = time.Now()
		a.addTransferLog(order, name, ttype, int(msg.State), reason, false)
		order.Update()
	}
	rolePid.Request(msg, rolePid)
}

func (a *RoleActor) withdrawSubmit(outOrder *data.WithdrawOrder, transfer bool) (*data.WithdrawResponse, error) {
	req := data.WithdrawRequest{
		OrderID:    outOrder.OrderID,
		Userid:     outOrder.Userid,
		RealName:   outOrder.RealName,
		Bank:       outOrder.Bank,
		BankNumber: outOrder.BankNumber,
		IFSC:       outOrder.IFSC,
		Amount:     outOrder.Amount,
		Transfer:   transfer,
		PayWay:     outOrder.PayWay,
	}
	strReq, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal err ,order:%s, user:%s,err:%v", outOrder.OrderID, outOrder.Userid, err)
	}
	resp, err1 := handler.DoHttpPostByJson(withdrawurl, strReq)
	if err1 != nil {
		return nil, fmt.Errorf("http post err ,order:%s, user:%s,err:%v", outOrder.OrderID, outOrder.Userid, err1)
	}
	res := new(data.WithdrawResponse)
	err2 := jsoniter.Unmarshal(resp, res)
	if err2 != nil {
		return nil, fmt.Errorf("unmarshal [WithdrawResponse] err ,order:%s, user:%s,err:%v", outOrder.OrderID, outOrder.Userid, err2)
	}
	return res, nil
}

// 转单
func (a *RoleActor) transferWithdrawOrder(order *data.WithdrawRecord, name string, ttype int) {
	if order.OrderStatus != data.DeliveryFailTransfer &&
		order.OrderStatus != data.WithdrawUrgentTransfer &&
		order.OrderStatus != data.WithdrawTimeoutTransfer &&
		order.OrderStatus != data.WithdrawFail &&
		order.OrderStatus != data.OrderFail &&
		order.OrderStatus != data.DeliveryFailRefund &&
		order.OrderStatus != data.WithdrawFailRefund &&
		order.OrderStatus != data.WithdrawFailTransfer {
		glog.Errorf("order status error, can't transfer. order %s, user:%s", order.OrderID, order.Userid)
		return
	}
	reason := order.GetTransferReason()
	go a.submitWithdrawOrder(order, true, name, reason, ttype)
}

// 回退订单
func (a *RoleActor) withdrawOrderBack(order *data.WithdrawRecord, status int) string {
	// if order.OrderStatus != data.WithdrawFail && order.OrderStatus != data.OrderFail && order.OrderStatus != data.Withdrawing {
	// if order.OrderStatus == data.WithdrawSuccess || order.OrderStatus == data.OrderSuccess || order.OrderStatus == data.WithdrawBack {
	// 	return "back fail, status, exception"
	// }
	if utils.SliceIn(order.OrderStatus,
		data.WithdrawSuccess,
		data.WithdrawBack,
		data.WithdrawFreeze,
		data.OrderSuccess,
		data.WithdrawRefunded) {
		return fmt.Sprintf("withdraw order back status error: %s, %d", order.OrderID, order.OrderStatus)
	}
	// 在线时
	msg3 := handler.ChangeCurrencyMsg(int64(order.Score+order.Commission), 0,
		0, 0, 0, int64(order.Score), int32(pb.LOG_TYPE63), order.Userid, "提现回退", "")
	if v, ok := a.roles[order.Userid]; ok {
		v.Pid.Tell(msg3)
		msg := new(pb.WithdrawLogStatus)
		msg.OrderId = order.OrderID
		msg.ShopId = order.ShopId
		msg.Status = data.WithdrawBack
		msg.Amount = -int32(order.Amount)
		v.Pid.Tell(msg)
	} else {
		user := a.getUserById(order.Userid)
		if user == nil {
			return fmt.Sprintf("user is nil, id:%s", order.OrderID)
		}
		log := user.WithdrawLogMap[order.OrderID]
		if log != nil {
			log.Status = data.WithdrawBack
		}

		if utils.Time2DayDate(order.Ctime.In(location)) == utils.Time2DayDate(time.Now().In(location)) {
			// 同一天才退次数
			if user.Vip.WithdrawCount > 0 {
				user.Vip.WithdrawCount--
				user.Vip.WithdrawAmount -= int(order.Amount)
			}
		}
		// user.WithDrawCountMap[order.ShopId]--
		user.CashOut -= int32(order.Amount)
		user.UpdateWithdrawLog()

		// 邮件
		if config.SettingIsOpen(4, data.MAILREPLY) {
			ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: status})
			if ntf != nil {
				user.UpdateFeedBack()
			}
		}

		// 银行信息错误的邮件
		// if status == data.WithdrawBackBankErr {
		// 	ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: data.WithdrawBackBankErr})
		// 	if ntf != nil {
		// 		user.UpdateFeedBack()
		// 	}
		// }
		// if _, ok := a.offline[order.OrderID]; !ok {
		// 	// 同步db
		// }
		// 离线时
		rolePid.Tell(msg3)
	}
	order.OrderStatus = int32(status)
	order.WithdrawTime = utils.BsonNow()
	if !order.Update() {
		return fmt.Sprintf("update order fail, id:%s", order.OrderID)
	}
	return ""
}

// 冻结订单
func (a *RoleActor) withdrawOrderFreeze(order *data.WithdrawRecord) string {
	if utils.SliceIn(order.OrderStatus,
		data.WithdrawSuccess,
		data.WithdrawBack,
		data.WithdrawFreeze,
		data.OrderSuccess) {
		return fmt.Sprintf("withdraw order freeze status error: %s, %d", order.OrderID, order.OrderStatus)
	}
	if v, ok := a.roles[order.Userid]; ok {
		msg := new(pb.WithdrawLogStatus)
		msg.OrderId = order.OrderID
		msg.ShopId = order.ShopId
		msg.Status = data.WithdrawFreeze
		msg.Amount = -int32(order.Amount)
		v.Pid.Tell(msg)
	} else {
		user := a.getUserById(order.Userid)
		if user == nil {
			return fmt.Sprintf("user is nil, id:%s", order.OrderID)
		}
		log := user.WithdrawLogMap[order.OrderID]
		if log != nil {
			log.Status = data.WithdrawSuccess // 冻结的订单显示成功
		}
		user.CashOut -= int32(order.Amount)
		if _, ok := a.offline[order.Userid]; !ok {
			// 同步db
			user.UpdateWithdrawLog()
		}
	}
	order.OrderStatus = data.WithdrawFreeze
	order.WithdrawTime = utils.BsonNow()
	if !order.Update() {
		return fmt.Sprintf("update order fail, id:%s", order.OrderID)
	}
	return ""
}

// 提现订单更新
func (a *RoleActor) withdrawOrderUpdate(arg *pb.PayOrderUpdate, ctx actor.Context) {
	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	rsp := new(pb.PayedOrderUpdate)
	order.Get()
	if order.OrderStatus == data.WithdrawSuccess {
		glog.Infof("order already success,id:%s, state:%d", order.OrderID, order.OrderStatus)
		rsp.Result = true
		ctx.Respond(rsp)
		return
	}

	if order.OrderStatus != data.Withdrawing && order.OrderStatus != data.WaitDelivery &&
		(arg.State == data.WithdrawBack || arg.State == data.WithdrawFreeze) { // 退款或者冻结
		// 不是审核中或者待派单就不处理了
		glog.Infof("repeat handle order,id:%s, state:%d", order.OrderID, order.OrderStatus)
		rsp.Result = true
		ctx.Respond(rsp)
		return
	}

	userState := arg.GetState()

	order.OrderStatus = arg.GetState()
	order.OutTradeNo = arg.OutId
	order.RefuseReason = arg.Err
	order.OutTradeStatus = arg.OutStatus
	order.Utime = utils.BsonNow()
	if arg.State == data.WithdrawSuccess {
		order.WithdrawTime = utils.BsonNow()
		order.RepeatTimes += 1
	}
	if arg.ChannelId != 0 {
		order.OutChannel = utils.String(arg.ChannelId)
	}
	// 添加转单记录
	a.addTransferLog(order, "", data.AutoTransfer, int(arg.State), int(arg.Reason), false)

	// switch arg.OutStatus {
	// case "3":
	// 	// 失败的订单，先改为处理中
	// 	userState = data.Withdrawing
	// }
	if !order.Update() {
		glog.Errorf("update withdraworder fail,id:%s, state:%d", order.OrderID, order.OrderStatus)
		return
	}

	// 失败转单
	if order.OrderStatus == data.WithdrawFailTransfer || order.OrderStatus == data.DeliveryFailTransfer {
		mq.NatsPublish(mq.TopicWithdrawFail, &pb.WithdrawTransfer{OrderId: order.OrderID})
	}

	// 订单更新
	if v, ok := a.roles[order.Userid]; ok {
		msg := new(pb.WithdrawLogStatus)
		msg.OrderId = arg.GetOrderId()
		msg.ShopId = order.ShopId
		msg.Status = userState
		msg.BankName = order.Blank
		v.Pid.Tell(msg)
	} else {
		user := a.getUserById(order.Userid)
		if user == nil {
			glog.Errorf("user is nil, orderid:%s", order.OrderID)
			rsp.Result = true
			ctx.Respond(rsp)
			return
		}
		log := user.WithdrawLogMap[order.OrderID]
		if log != nil {
			log.Status = int(userState)
		}
		if config.SettingIsOpen(4, data.MAILREPLY) {
			// 邮件
			ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: log.Status})
			if ntf != nil {
				user.UpdateFeedBack()
			}
		}
		// 同步db
		user.UpdateWithdrawLog()
		// 提现次数记录
		a.SavePKWithdrawInfo(user, order.Blank, int(arg.State))
	}
	ctx.Respond(rsp)
}

func (a *RoleActor) RepeatWithdraw(ctx actor.Context) {
	msg := ctx.Message()
	// 重复支付
	arg := msg.(*pb.RepeatWithdraw)
	glog.Debugf("RepeatWithdraw %#v", arg)

	rsp := new(pb.RepeatWithdrawed)
	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	defer ctx.Respond(rsp)

	if order.OrderID == "" {
		glog.Error("repeat withdraw no found order,id:%s", arg.OrderId)
		return
	}

	// 记录日志
	update := false
	for _, t := range order.TransferDetail {
		if t.ChannelId != arg.ChannelId {
			continue
		}
		update = true

		if t.Status == int(arg.Status) {
			glog.Errorf("repeat withdraw order,id:%s, state:%d", order.OrderID, arg.Status)
			break
		}
		t.Status = int(arg.Status)
		t.Utime = time.Now().Unix()
		break
	}

	if !update {
		de := &data.TransferOrderDetail{
			ChannelId: arg.ChannelId,
			TType:     data.AutoTransfer,
			Status:    int(arg.Status),
			Ctime:     time.Now().Unix(),
		}
		order.TransferDetail = append(order.TransferDetail, de)
	}

	if arg.Status == data.WithdrawSuccess {
		order.AddRepeatTimes()
	} else {
		order.UpdateTransferLog()
	}
}

func (a *RoleActor) addTransferLog(order *data.WithdrawRecord, name string, ttype, status, reason int, update bool) {
	c, _ := strconv.Atoi(order.OutChannel)
	channelId := uint32(c)
	// 添加转单记录
	if channelId == 0 {
		channel, _ := strconv.ParseUint(order.OutChannel, 10, 64)
		channelId = uint32(channel)
	}

	insert := true
	for _, t := range order.TransferDetail {
		if t.ChannelId != channelId {
			continue
		}
		insert = false
		t.Status = status
		// t.Reason = reason
		t.Utime = time.Now().Unix()
	}

	if insert {
		de := &data.TransferOrderDetail{
			ChannelId: channelId,
			TType:     ttype,
			Status:    status,
			UserName:  name,
			Ctime:     time.Now().Unix(),
			Reason:    reason,
		}
		order.TransferDetail = append(order.TransferDetail, de)
	}

	if update {
		order.UpdateTransferLog()
	}
}

// 提现记录超时退款请求(用户申请提现退款)
func (a *RoleActor) WithdrawRecordRefundReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordRefundReq)
	rsp := &pb.WithdrawRecordRefundRsp{}
	glog.Debugf("WithdrawRecordRefundReq %#v", arg)

	defer ctx.Respond(rsp)

	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	if order.OrderID == "" {
		glog.Errorf("withdraw order no found, id:%s", arg.OrderId)
		rsp.Error = pb.Failed
		return
	}

	// 订单处理中或处理完毕不可退款
	if order.OrderStatus != data.Withdrawing || order.Userid != arg.Userid {
		rsp.Error = pb.Failed
		return
	}

	now := time.Now().In(location)
	ctime := order.Ctime.In(location)
	refundTime := table.GetTables().PaymentTable.Get().RefundTime[arg.RegistArea]
	if now.Before(ctime.Add(time.Duration(refundTime) * time.Hour)) {
		// 可退款时间未到
		rsp.Error = pb.Failed
		return
	}

	// 补偿金额
	refundCompensateRate := table.GetTables().PaymentTable.Get().RefundCompensateRate[arg.RegistArea]
	refundCompensateAmount := table.GetTables().PaymentTable.Get().RefundCompensateAmount[arg.RegistArea]
	refundCompensateType := table.GetTables().PaymentTable.Get().RefundCompensateType[arg.RegistArea]
	refundCompensate := int64(math.Floor(float64(order.Amount)*(float64(refundCompensateRate)/10000))) + int64(refundCompensateAmount)
	var msgRefundCompensate *pb.ChangeCurrency
	if refundCompensate > 0 {
		// 发送补偿 1代表bonus,2代表cash,3代表withdrawable
		var bonus, diamond, outDiamond int64
		switch refundCompensateType {
		case 1:
			bonus = refundCompensate
		case 2:
			diamond = refundCompensate
		case 3:
			diamond = refundCompensate
			outDiamond = refundCompensate
		}
		msgRefundCompensate = handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
			int32(pb.LOG_TYPE137), order.Userid, "提现申请退款补偿", order.OrderID)
	}
	// 退款
	// 在线时
	msg3 := handler.ChangeCurrencyMsg(int64(order.Score+order.Commission), 0,
		0, 0, 0, int64(order.Score), int32(pb.LOG_TYPE136), order.Userid, "申请提现退款", order.OrderID)
	if v, ok := a.roles[order.Userid]; ok {
		v.Pid.Tell(msg3)
		if msgRefundCompensate != nil {
			v.Pid.Tell(msgRefundCompensate)
		}
		msg := new(pb.WithdrawLogStatus)
		msg.OrderId = order.OrderID
		msg.ShopId = order.ShopId
		msg.Status = data.WithdrawRefunded
		msg.Amount = -int32(order.Amount)
		v.Pid.Tell(msg)
	} else {
		user := a.getUserById(order.Userid)
		if user == nil {
			glog.Errorf("WithdrawRecordRefundReq user is nil, userid=%s, order=%s", order.Userid, order.OrderID)
			rsp.Error = pb.Failed
			return
		}
		log := user.WithdrawLogMap[order.OrderID]
		if log != nil {
			log.Status = data.WithdrawRefunded
		}

		if utils.Time2DayDate(order.Ctime.In(location)) == utils.Time2DayDate(time.Now().In(location)) {
			// 同一天才退次数
			if user.Vip.WithdrawCount > 0 {
				user.Vip.WithdrawCount--
				user.Vip.WithdrawAmount -= int(order.Amount)
			}
		}
		user.CashOut -= int32(order.Amount)
		user.UpdateWithdrawLog()

		// 邮件
		if config.SettingIsOpen(4, data.MAILREPLY) {
			ntf := event.Event(user, event.WITHDRAW_Mail, &event.WithdrawMailEvent{Status: data.WithdrawRefunded})
			if ntf != nil {
				user.UpdateFeedBack()
			}
		}

		// 离线时
		rolePid.Tell(msg3)
		if msgRefundCompensate != nil {
			rolePid.Tell(msgRefundCompensate)
		}
	}
	order.OrderStatus = int32(data.WithdrawRefunded)
	order.WithdrawTime = utils.BsonNow()
	order.RefundCompensate = refundCompensate
	order.RefundCompensateType = refundCompensateType
	if !order.Update() {
		glog.Errorf("update order fail, id:%s", order.OrderID)
		rsp.Error = pb.Failed
		return
	}
}

// 充值记录utr回填请求
func (a *RoleActor) PaymentRecordUTRReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PaymentRecordUTRReq)
	rsp := &pb.PaymentRecordUTRRsp{}
	glog.Debugf("PaymentRecordUTRReq %#v", arg)

	defer ctx.Respond(rsp)

	order := &data.TradeRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	if order.OrderID == "" {
		glog.Error("pay order no found, id:%s", arg.OrderId)
		rsp.Error = pb.Failed
		return
	}

	if order.OrderStatus != data.Tradeing || order.Userid != arg.Userid {
		// 重复处理的订单
		rsp.Error = pb.Failed
		return
	}

	now := time.Now().In(location)
	ctime := order.Ctime.In(location)
	// 订单已超时
	payingTimeLimit := table.GetTables().PaymentTable.Get().PayingTimeLimit
	if now.After(ctime.Add(time.Duration(payingTimeLimit) * time.Minute)) {
		rsp.Error = pb.Failed
		return
	}
	if order.LastUtr == arg.UserUTR {
		return
	}

	utrCountLimit := table.GetTables().PaymentTable.Get().UtrCountLimit
	order.LastUtr = arg.UserUTR
	order.LastUtrTime = now.Unix()

	// 提交一样的覆盖
	var utrs []data.TradeRecordUTR
	for _, utr := range order.UserUtrs {
		if utr.UTR != arg.UserUTR {
			utrs = append(utrs, utr)
		}
	}
	utrs = append(utrs, data.TradeRecordUTR{UTR: arg.UserUTR, Time: now.Unix()})
	if len(utrs) > int(utrCountLimit) {
		utrs = utrs[len(utrs)-int(utrCountLimit):]
	}
	order.UserUtrs = utrs

	order.Upsert()
}

// 提现记录超时继续等待请求
func (a *RoleActor) WithdrawRecordKeepWaitingReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordKeepWaitingReq)
	rsp := &pb.WithdrawRecordKeepWaitingRsp{}
	glog.Debugf("WithdrawRecordKeepWaitingReq %#v", arg)

	defer ctx.Respond(rsp)

	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	if order.OrderID == "" {
		glog.Errorf("repeat withdraw no found order, id:%s", arg.OrderId)
		rsp.Error = pb.Failed
		return
	}

	// 订单处理中或处理完毕
	if order.OrderStatus != data.Withdrawing || order.Userid != arg.Userid {
		rsp.Error = pb.Failed
		return
	}
	if order.KeepWaitTime > 0 {
		return
	}

	order.KeepWaitTime = time.Now().Unix()
	order.Upsert()
}

// 提现超时赔付领取
func (a *RoleActor) WithdrawRecordOvertimePayTackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawRecordOvertimePayTackReq)
	rsp := &pb.WithdrawRecordOvertimePayTackRsp{}
	glog.Debugf("WithdrawRecordOvertimePayTackReq %#v", arg)

	defer ctx.Respond(rsp)

	// 提现保障时效（小时）
	withdrawSafeguardTime := table.GetTables().PaymentTable.Get().WithdrawSafeguardTime[arg.RegistArea]
	// 超时赔付金
	withdrawTimeoutPayment := table.GetTables().PaymentTable.Get().WithdrawTimeoutPayment[arg.RegistArea]
	withdrawOvertimeBeginDateS := table.GetTables().PaymentTable.Get().WithdrawOvertimeBeginDate
	withdrawOvertimeBeginDate, overtimeErr := time.ParseInLocation(utils.FORMAT_DATE, withdrawOvertimeBeginDateS, location)
	if overtimeErr != nil {
		glog.Errorf("withdrawOvertimeBeginDate parse error: %s, %v", withdrawOvertimeBeginDate, overtimeErr)
		rsp.Error = pb.Failed
		return
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

	// 超时赔偿活动关闭
	if !(overtimePay > 0 && withdrawSafeguardTime > 0) {
		rsp.Error = pb.Failed
		return
	}

	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	if order.OrderID == "" {
		glog.Error("repeat withdraw no found order, id:%s", arg.OrderId)
		rsp.Error = pb.Failed
		return
	}

	if order.Userid != arg.Userid { // order.OrderStatus != data.Withdrawing ||
		rsp.Error = pb.Failed
		return
	}

	now := time.Now().In(location)
	ctime := order.Ctime.In(location)

	// 活动之前的订单不处理
	if ctime.Before(withdrawOvertimeBeginDate) {
		rsp.Error = pb.Failed
		return
	}

	// 超时赔付计算
	finishTime := now // 完结时间
	if !utils.SliceIn(order.OrderStatus, data.Withdrawing, data.OrderSuccess) {
		finishTime = order.WithdrawTime.In(location)
	}
	passHours := finishTime.Sub(ctime).Hours()
	// 未到赔付时间
	if passHours <= 0 {
		rsp.Error = pb.Failed
		return
	}

	// 计算赔偿次数
	overtimePayTimes := int32(math.Floor(passHours / float64(withdrawSafeguardTime)))
	// 赔偿金
	overtimePays := overtimePay * overtimePayTimes
	// 可领取的赔偿金
	hasOvertimePay := int64(overtimePays) - order.OvertimePayTacks
	if hasOvertimePay <= 0 {
		rsp.Error = pb.Failed
		return
	}

	// 发送补偿 1代表bonus,2代表cash,3代表withdrawable
	var bonus, diamond, outDiamond int64
	switch overtimePayType {
	case 1:
		bonus = hasOvertimePay
	case 2:
		diamond = hasOvertimePay
	case 3:
		diamond = hasOvertimePay
		outDiamond = hasOvertimePay
	}

	// 更新领取情况
	order.OvertimePayTacks += hasOvertimePay
	order.OvertimePayTackTimes++
	order.Upsert()

	msgOvertimePay := handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
		int32(pb.LOG_TYPE138), order.Userid, "提现超时赔付领取", order.OrderID)
	// 在线时
	if v, ok := a.roles[order.Userid]; ok {
		v.Pid.Tell(msgOvertimePay)
	} else {
		// 离线时
		rolePid.Tell(msgOvertimePay)
	}
}

// 订单回填UTR额外赠送
func (a *RoleActor) PayUTRBackReword(ctx actor.Context) {
	arg := ctx.Message().(*pb.PayUTRBackReword)
	if arg.LastUtr == "" {
		return
	}

	utrBackRewordRate := table.GetTables().PaymentTable.Get().UtrBackRewordRate[arg.RegistArea]
	utrBackRewordType := table.GetTables().PaymentTable.Get().UtrBackRewordType[arg.RegistArea]
	utrGive := int64(float64(arg.Amount) * (float64(utrBackRewordRate) / 10000))
	if utrGive <= 0 {
		return
	}

	// 发送补偿 1代表bonus,2代表cash,3代表withdrawable
	var bonus, diamond, outDiamond int64
	switch utrBackRewordType {
	case 1:
		bonus = utrGive
	case 2:
		diamond = utrGive
	case 3:
		diamond = utrGive
		outDiamond = utrGive
	}

	msg := handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
		int32(pb.LOG_TYPE139), arg.Userid, "回填UTR额外赠送", arg.OrderId)
	// 在线时
	if v, ok := a.roles[arg.Userid]; ok {
		v.Pid.Tell(msg)
	} else {
		// 离线时
		rolePid.Tell(msg)
	}

	glog.Infof("回填utr赠送: %s, %s, %d, %d, %d, %d", arg.Userid, arg.OrderId, arg.Amount, bonus, diamond, outDiamond)
}

func (a *RoleActor) UploadUtrReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UploadUtrReq)
	glog.Debugf("UploadUtrReq %#v", arg)

	utr := &data.Utr{
		Uid:      arg.Uid,
		Orderid:  arg.OrderId,
		FileName: arg.FileName,
		Status:   int(data.UtrStatusPending),
	}

	if !utr.Save() {
		glog.Errorf("UploadUtrReq save utr error: %v", err)
		return
	}

	order := &data.TradeRecord{OrderID: arg.OrderId}
	order.Get()
	if order.OrderID == "" {
		glog.Errorf("UploadUtrReq order not found, id:%s", arg.OrderId)
		return
	}
	arg.ChannelId = int32(order.ChannelId)

	if err := mq.NatsPublish(mq.TopicUtrUpload, arg); err != nil {
		glog.Errorf("UploadUtrReq publish error: %v", err)
	}
}

func (a *RoleActor) UploadUtrRsp(ctx actor.Context) {
	arg := ctx.Message().(*pb.UploadUtrRsp)
	glog.Debugf("UploadUtrRsp %#v", arg)

	utr := &data.Utr{
		Uid:      arg.Uid,
		Orderid:  arg.OrderId,
		FileName: arg.FileName,
		Status:   int(data.UtrStatusPending),
	}
	utr.Get()

	if utr.Id == "" {
		glog.Errorf("UploadUtrRsp utr not found, id:%s", arg.OrderId)
		return
	}

	utr.Status = int(arg.Status)
	utr.RefNo = arg.RefNo
	utr.RefType = arg.RefType
	utr.PayTime = arg.PayTime
	utr.Amount = arg.Amount
	utr.Recipient = arg.Recipient
	// 重复上传同一个
	if utr.Has() {
		return
	}

	// if err := utr.Update(); err != nil {
	// 	glog.Errorf("UploadUtrRsp update utr error: %v", err)
	// }
	defer utr.Update()

	if r, ok := a.roles[utr.Uid]; ok {
		r.Pid.Tell(&pb.OrderUtrResultNtf{OrderId: arg.OrderId, Status: arg.Status})
		utr.IsNotify = true
	}
}

func (a *RoleActor) HasFailUtr(ctx actor.Context) {
	arg := ctx.Message().(*pb.HasFailUtr)
	glog.Debugf("HasFailUtr %#v", arg)
	utr := &data.Utr{
		Uid: arg.Uid,
	}

	list := utr.ListByQ(bson.M{
		"uid":       utr.Uid,
		"is_notify": utr.IsNotify,
		"status":    2,
	})

	defer utr.UpdateNotify()

	rsp := &pb.HasFailUtred{Has: len(list) > 0}
	ctx.Respond(rsp)
}

// 自动转单
// func (a *RoleActor) AutoTransferOrder(reason int) {
// 	// 开关
// 	var timeout int32
// 	switch reason {
// 	case data.TimeoutTransfer:
// 		timeout = table.GetTables().PaymentTable.Get().WithdrawReviewTimeOut
// 	case data.TimeoutUrgentTransfer:
// 		timeout = table.GetTables().PaymentTable.Get().WithdrawReviewTimeOutRapid
// 	default:
// 		return
// 	}
// 	if timeout == 0 {
// 		return
// 	}
// 	// 超时转单
// 	// 查询提单成功的订单(超过3天不查了)
// 	startTime := time.Now().UTC().Add(-time.Duration(timeout) * time.Minute)
// 	endTime := time.Now().UTC().Add(-time.Duration(3*24) * time.Hour)
// 	query := bson.M{
// 		"order_status": data.OrderSuccess,
// 		"e_time": bson.M{"$lte": startTime,
// 			"$gt": endTime,
// 		},
// 	}
// 	list := data.WithdrawRecordListByQ(query)
// 	tgMsg := make([]string, 0)
// 	for _, order := range list {
// 		// 检查当前用户有没有转单补贴金
// 		user := a.getUserById(order.Userid)
// 		if user == nil {
// 			glog.Warningf("用户%s 不存在, 不进行转单", order.Userid)
// 			continue
// 		}
// 		if user.TransferCash >= int64(order.Amount) {
// 			glog.Warningf("用户%s 转单补贴金已用完, 不进行转单", order.Userid)
// 			continue
// 		}
// 		exist := false
// 		for _, detail := range order.TransferDetail {
// 			if detail.Reason == reason {
// 				exist = true
// 				break
// 			}
// 			if data.TimeoutTransfer == reason &&
// 				(detail.Reason == data.TimeoutTransfer ||
// 					detail.Reason == data.TimeoutUrgentTransfer ||
// 					detail.Reason == data.FailedTransfer) {
// 				exist = true
// 				break
// 			}
// 		}
// 		if exist {
// 			glog.Warningf("用户%s 已存在转单记录, 不进行转单", order.Userid)
// 			continue
// 		}
// 		// 转单
// 		a.transferWithdrawOrder(order, "System", reason, data.AutoTransfer)
// 		tgMsg = append(tgMsg, order.OrderID)
// 		// 扣转单补贴金
// 		user.TransferCash -= int64(order.Score)
// 		user.UpdateTransferCash()
// 	}
// 	// tg消息
// 	if len(tgMsg) > 0 {
// 		// msg := fmt.Sprintf("超时转单:\n %s", strings.Join(tgMsg, "\n"))
// 		msg := &pb.AlertorMail{
// 			Subject: "超时转单",
// 			Message: strings.Join(tgMsg, "\n"),
// 		}
// 		mq.NatsPublish(mq.TopicAlertEmail, msg)
// 	}
// }

func (a *RoleActor) WithdrawTransfer(ctx actor.Context) {
	arg := ctx.Message().(*pb.WithdrawTransfer)
	glog.Debugf("WithdrawTransfer %#v", arg)

	order := &data.WithdrawRecord{
		OrderID: arg.OrderId,
	}
	order.Get()

	if order.OrderID == "" {
		glog.Errorf("WithdrawTransfer order not found, id:%s", arg.OrderId)
		return
	}

	// 派单
	if arg.Ttype == data.NoTransfer && order.OrderStatus == data.WaitDelivery {
		go a.submitWithdrawOrder(order, false, arg.Name, data.NormalTransfer, data.NoTransfer)
		return
	}

	// if order.OrderStatus != data.OrderSuccess &&
	// 	order.OrderStatus != data.OrderFail &&
	// 	order.OrderStatus != data.WithdrawFail &&
	// 	order.OrderStatus != data.WithdrawFailTransfer &&
	// 	order.OrderStatus != data.DeliveryFailTransfer &&
	// 	order.OrderStatus != data.WithdrawTimeoutTransfer &&
	// 	order.OrderStatus != data.WithdrawUrgentTransfer {
	// 	glog.Errorf("WithdrawTransfer order status not support, id:%s", arg.OrderId)
	// 	return
	// }

	// 转单
	a.transferWithdrawOrder(order, arg.Name, int(arg.Ttype))
}

func (a *RoleActor) checkTransferCash(order *data.WithdrawRecord) bool {
	if order.OrderStatus != data.WithdrawUrgentTransfer &&
		order.OrderStatus != data.WithdrawTimeoutTransfer {
		return true
	}
	transferCashStr, _ := myredis.Redis().Get(context.Background(), data.TransferCashKey).Result()
	transferCash, _ := strconv.ParseInt(transferCashStr, 10, 64)
	if transferCash < int64(order.Amount) {
		return false
	}

	// transferCash -= int64(order.Amount)
	myredis.Redis().IncrBy(context.Background(), data.TransferCashKey, -int64(order.Amount))
	return true
}
