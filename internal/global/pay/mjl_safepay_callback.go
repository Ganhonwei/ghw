package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/safepay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"runtime/debug"

	"github.com/valyala/fasthttp"
)

func safePayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	glog.Infof("safepay clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()

	paramMap, err1 := safepay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	sign := safepay.PaySign(paramMap, safepay.Config.Key)
	if sign != outSign {
		glog.Errorf("[safepay] order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		safePayIncome(ctx, paramMap)
	case 2:
		safePayWithdraw(ctx, paramMap)
	}
}

// safePay充值回调
func safePayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_no"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["order_realityamount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[safepay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "success" {
		status = 3
		glog.Errorf("[safepay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[safepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_SAFEPAY {
		strmsg := "[safepay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "ok")
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
		strmsg := "[safepay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "ok")
		return
	}
	fmt.Fprint(ctx, "ok")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// safepay提现回调
func safePayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_no"]
	order_status := paramMap["result"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["order_realityamount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[safepay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "success" {
		glog.Errorf("[safepay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[safepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_SAFEPAY, status) {
		glog.Warning("[safepay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "ok")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[safepay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "ok")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[safepay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[safepay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "ok")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "ok")

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
