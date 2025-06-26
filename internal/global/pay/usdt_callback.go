package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/internal/global/pay/usdtpay"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/valyala/fasthttp"
)

// usdt支付回调(1：代收  2:代付)
func usdtpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	glog.Infof("usdt notify clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	paramMap, err1 := usdtpay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["signature"]
	// 验证签名
	delete(paramMap, "signature")

	var key string = usdtpay.Config.Md5Key
	var orderid string = paramMap["order_id"].(string)
	var businessid string = ""
	sign := usdtpay.PaySign(paramMap, key)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "usdt回调", "[usdt] order callback verify sign fail", sign, rtype)

		glog.Errorf("[usdt]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		usdtpayIncome(ctx, paramMap, businessid)
	case 2:
		usdtpayWithdraw(ctx, paramMap)
	}
}

// 充值回调
func usdtpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]any, businessid string) {
	status := 0
	orderid := paramMap["order_id"].(string)
	order_status := paramMap["status"].(string)
	order_amount := paramMap["amount"].(float64)
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[usdtpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "usdtpay回调", "usdtpay回调代收验签通过", string(jsonStr), 1)

	if order_status != "2" {
		status = 3
		glog.Errorf("[usdtpay] order %s callback fail, status: %s", orderid, order_status)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "usdtpay回调", "usdtpay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[usdtpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "usdtpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		strmsg := "[usdtpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "usdtpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "ok")
		return
	}

	info.OutTradeStatus = order_status
	milliseconds := utils.BsonNow().UnixNano() / 1e6
	info.PayTime = milliseconds
	info.CallBackMsg = str
	info.OrderStatus = int32(status)

	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[usdtpay] order " + orderid + "Update order error,msg:" + err.Error()
		PayPointRecord(businessid, "usdtpay回调", strmsg, orderid, 1)
	}

	// 通知服务器
	amount := order_amount // 金额(元)
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
		strmsg := "[usdtpay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "usdtpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "ok")
		return
	}
	strmsg := "[usdtpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "usdtpay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "ok")

	if status == 2 {
		// 订单记录
		OrederRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 提现回调
func usdtpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]any) {
	status := 0
	orderid := paramMap["order_id"].(string)
	order_status := paramMap["status"].(string)
	order_amount := paramMap["amount"].(float64)
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[usdtpay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "3" {
		glog.Errorf("[usdtpay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[usdtpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.USDT, status) {
		glog.Warning("[usdtpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "ok")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[usdtpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "ok")
		return
	}

	info.OutTradeStatus = order_status
	milliseconds := utils.BsonNow().UnixNano() / 1e6
	info.PayTime = milliseconds
	// info.PayTime, _ = strconv.ParseInt(order_time, 10, 64)
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[usdtpay] order " + orderid + "Update order error,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount := order_amount // 金额(元)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = int(amount * 100)
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[usdtpay] order " + orderid + " withdraw callback notify fail,err:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
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
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
