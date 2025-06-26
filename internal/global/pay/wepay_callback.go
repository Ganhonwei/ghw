package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/internal/global/pay/wepay"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strconv"

	"github.com/valyala/fasthttp"
)

// wepay支付回调(1：代收  2:代付)
func wepayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
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
	glog.Infof("wepay notify clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	paramMap, err1 := wepay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")

	var key string = wepay.Config.Md5Key
	var orderid string = paramMap["orderNo"]
	var businessid string = ""
	sign := wepay.PaySign(paramMap, key)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "wepay回调", "[wepay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[wepay]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		wepayIncome(ctx, paramMap, businessid)
	case 2:
		wepayWithdraw(ctx, paramMap)
	}
}

// 充值回调
func wepayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["payStatus"]
	order_time := paramMap["payTime"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[wepay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "wepay回调", "wepay回调代收验签通过", string(jsonStr), 1)

	if order_status != "1" {
		status = 3
		glog.Errorf("[wepay] order %s callback fail, status: %s", orderid, order_status)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "wepay回调", "wepay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[wepay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "wepay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.WEPAY {
		strmsg := "[wepay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "wepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	// layout := "2006-01-02 15:04:05"
	if t, err := strconv.ParseInt(order_time, 10, 64); err == nil {
		info.PayTime = t
	} else {
		glog.Errorf("wepay paytime format error: %s", order_time)
	}
	// info.PayTime, _ = strconv.ParseInt(order_time, 10, 64)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[wepay] order " + orderid + "Update order error,msg:" + err.Error()
		PayPointRecord(businessid, "wepay回调", strmsg, orderid, 1)
	}

	// 通知服务器
	amount, err := strconv.ParseFloat(order_amount, 64) // 金额(元)
	if err != nil {
		glog.Errorf("wepay order amount parse error: %s", order_amount)
	}
	// amount, _ := strconv.Atoi(order_amount)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = int(amount * 100)
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[wepay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "wepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "success")
		return
	}
	strmsg := "[wepay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "wepay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "success")

	if status == 2 {
		// 订单记录
		OrederRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 提现回调
func wepayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["payStatus"]
	order_time := paramMap["payTime"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[wepay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "1" {
		glog.Errorf("[wepay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[wepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.WEPAY, status) {
		glog.Warning("[wepay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[wepay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	// layout := "2006-01-02 15:04:05"
	if t, err := strconv.ParseInt(order_time, 10, 64); err == nil {
		info.PayTime = t
	} else {
		glog.Errorf("wepay paytime format error: %s", order_time)
	}
	// info.PayTime, _ = strconv.ParseInt(order_time, 10, 64)
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[wepay] order " + orderid + "Update order error,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount, err := strconv.ParseFloat(order_amount, 64) // 金额(元)
	if err != nil {
		glog.Errorf("wepay withdraw order amount parse error: %s", order_amount)
	}

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = int(amount * 100)
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[wepay] order " + orderid + " withdraw callback notify fail,err:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
		fmt.Fprint(ctx, "success")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "success")

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
