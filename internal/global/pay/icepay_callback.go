package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/icepay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strconv"

	"github.com/valyala/fasthttp"
)

// icepay支付回调(1：代收  2:代付)
func icepayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	var paramMap map[string]string
	var err1 error
	if rtype == 1 {
		paramMap, err1 = icepay.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = icepay.Config.ParsePayResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")

	paramMap2 := make(map[string]string)
	paramMap2["amount"] = paramMap["amount"]
	paramMap2["merchantId"] = paramMap["merchantId"]
	paramMap2["orderId"] = paramMap["orderId"]
	paramMap2["timestamp"] = paramMap["timestamp"]

	var key string = ""
	var orderid string = ""
	var businessid string = ""
	switch rtype {
	case 1:
		orderid = paramMap["orderId"]
		key = icepay.Config.Md5Key
	case 2:
		orderid = paramMap["orderId"]
		key = icepay.Config.Md5Key
	}
	sign := icepay.PaySign(paramMap2, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "icepay回调", "[icepay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[icepay]order %s verify sign fail", paramMap["orderId"])
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		icepayIncome(ctx, paramMap, businessid)
	case 2:
		icepayWithdraw(ctx, paramMap)
	}
}

// 充值回调
func icepayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	order_time := paramMap["timestamp"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[icepay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "icepay回调", "icepay回调代收验签通过", string(jsonStr), 1)

	if order_status != "1" {
		status = 3
		glog.Errorf("[icepay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "icepay回调", "icepay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[icepay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "icepay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.ICEPAY {
		strmsg := "[icepay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "icepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "OK")
		return
	}
	// timeStr := order_time                   // 时间字符串
	// layout := "Mon Jan 2 15:04:05 MST 2006" // 时间字符串的格式

	// t, err := time.Parse(layout, timeStr)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " payment time conversion failed,msg:" + err.Error()
	// 	PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	// }
	// milliseconds := t.UnixNano() / int64(time.Millisecond)

	info.OutTradeStatus = order_status
	info.PayTime, _ = strconv.ParseInt(order_time, 10, 64)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[icepay] order " + orderid + "Update order error,msg:" + err.Error()
		PayPointRecord(businessid, "icepay回调", strmsg, orderid, 1)
	}

	// 通知服务器
	// f, err := strconv.ParseFloat(order_amount, 64)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
	// 	PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	// }

	amount, _ := strconv.Atoi(order_amount)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp, _ = strconv.ParseInt(order_time, 10, 64)
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[icepay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "icepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "OK")
		return
	}
	strmsg := "[icepay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "icepay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "OK")

	if status == 2 {
		// 订单记录
		OrederRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 提现回调
func icepayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	order_time := paramMap["timestamp"]
	order_amount := paramMap["amount"]
	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[icepay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "1" {
		glog.Errorf("[icepay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[icepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.ICEPAY, status) {
		glog.Warning("[icepay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[icepay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "OK")
		return
	}
	// layout := "Mon Jan 2 15:04:05 MST 2006" // 时间字符串的格式

	// t, err := time.Parse(layout, order_time)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " payment time conversion failed,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	// }
	// milliseconds := t.UnixNano() / int64(time.Millisecond)

	info.OutTradeStatus = order_status
	info.PayTime, _ = strconv.ParseInt(order_time, 10, 64)
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// info.RefuseReason = msg
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[icepay] order " + orderid + "Update order error,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 通知服务器
	// f, err := strconv.ParseFloat(order_amount, 64)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	// }

	// i := int(f)
	amount, _ := strconv.Atoi(order_amount)

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp, _ = strconv.ParseInt(order_time, 10, 64)
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[icepay] order " + orderid + " withdraw callback notify fail,err:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
		fmt.Fprint(ctx, "OK")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "OK")

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
