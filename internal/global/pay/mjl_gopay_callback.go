package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/gopay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"runtime/debug"

	"github.com/valyala/fasthttp"
)

func goPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			debug.PrintStack()
		}
	}()
	ctx.Response.Header.Set("Content-type", "text/plain")
	switch string(ctx.Method()) {
	case "POST":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("gopay clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()

	paramMap, err1 := gopay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	sign := gopay.PaySign(paramMap, gopay.Config.Key)
	if sign != outSign {
		glog.Errorf("[gopay] order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		goPayIncome(ctx, paramMap)
	case 2:
		goPayWithdraw(ctx, paramMap)
	}
}

// safePay充值回调
func goPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]any) {
	status := 0
	orderid := paramMap["orderId"].(string)
	order_status := fmt.Sprint(paramMap["status"])
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[gopay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "1" {
		status = 3
		glog.Errorf("[gopay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[gopay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_GOPAY {
		strmsg := "[gopay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
	}

	amount := order_amount * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	// order.BusinessID = businessid
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[gopay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// gopay提现回调
func goPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]any) {
	status := 0
	orderid := paramMap["orderId"].(string)
	order_status := fmt.Sprint(paramMap["status"])
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[gopay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		glog.Errorf("[gopay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[gopay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_GOPAY, status) {
		glog.Warning("[gopay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[gopay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[gopay] order " + orderid + "Update order error,msg:" + err1.Error()
		glog.Errorf(strmsg)
	}

	amount := order_amount * 100

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[gopay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "SUCCESS")

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
