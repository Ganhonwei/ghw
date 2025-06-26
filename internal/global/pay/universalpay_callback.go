package pay

import (
	"crypto"
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/internal/global/pay/universalpay"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"net/http"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

// universalpay支付回调(1：代收  2:代付)
func universalpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	ctx.Response.Header.Set("Content-type", "text/plain")
	switch string(ctx.Method()) {
	case "POST":
	default:
		fmt.Fprint(ctx, "failure")
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	var paramMap map[string]string
	var err1 error
	if rtype == 1 {
		paramMap, err1 = universalpay.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = universalpay.Config.ParseWithdrawResult(body)
	}
	if err1 != nil {
		fmt.Fprint(ctx, err1.Error())
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}
	glog.Debugf("body %v", paramMap)
	serviceCode := string(ctx.Request.Header.Peek("X-SERVICE-CODE"))
	outSign := string(ctx.Request.Header.Peek("X-SIGN"))
	glog.Infof("universalpay notice serviceCode: %s", serviceCode)
	// 验证签名
	orderid := paramMap["orderNo"]
	businessid := service.GetOrderById(orderid, rtype)
	if !universalpay.Config.VerifySing(outSign, universalpay.SignContent(paramMap), crypto.SHA512) {
		// 签名验证失败
		PayPointRecord(businessid, "universalpay回调", "[universalpay] order callback verify sign fail", outSign, rtype)

		glog.Errorf("[universalpay] order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "FAILURE")
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}
	switch rtype {
	case 1:
		universalpayIncome(ctx, paramMap, businessid)
	case 2:
		universalpayWithdraw(ctx, paramMap)
	}
}

// universalpay充值回调
func universalpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["status"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[universalpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "universalpay回调", "universalpay回调代收验签通过", string(jsonStr), 1)
	if order_status != "1" {
		status = 3
		glog.Errorf("[universalpay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "universalpay回调", "universalpay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[universalpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.UNIVERSALPAY {
		strmsg := "[universalpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[universalpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SUCCESS")
		ctx.Response.SetStatusCode(http.StatusOK)
		return
	}
	// timeStr := order_time      // 时间字符串
	// layout := "20060102150405" // 时间字符串的格式
	// loc, _ := time.LoadLocation("Asia/Shanghai")
	// t, err := time.ParseInLocation(layout, timeStr, loc)
	payTime := time.Now().UnixMilli()
	if err != nil {
		strmsg := "[universalpay] order " + orderid + " payment time conversion failed, time:" + fmt.Sprint(payTime) + ", msg:" + err.Error()
		PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	info.OutTradeStatus = order_status
	info.PayTime = payTime
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[universalpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	// f, err := strconv.ParseFloat(order_amount, 64)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
	// 	PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	// }

	f, _ := strconv.ParseFloat(order_amount, 64)
	amount := int(f * 100)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = payTime
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[universalpay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		ctx.Response.SetStatusCode(http.StatusOK)
		return
	}
	strmsg := "[universalpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "universalpay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")
	ctx.Response.SetStatusCode(http.StatusOK)

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// universalpay提现回调
func universalpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["status"]
	order_amount := paramMap["amount"]
	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[universalpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		glog.Errorf("[universalpay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[universalpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.UNIVERSALPAY, status) {
		glog.Warning("[universalpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		ctx.Response.SetStatusCode(http.StatusOK)
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[universalpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		ctx.Response.SetStatusCode(http.StatusOK)
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = time.Now().UnixMilli()
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// info.RefuseReason = msg

	// if err1 != nil {
	// 	strmsg := "[pay1916] order " + orderid + "Update order error,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	// f, err := strconv.ParseFloat(order_amount, 64)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	// }

	f, _ := strconv.ParseFloat(order_amount, 64)
	amount := int(f * 100)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[universalpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		ctx.Response.SetStatusCode(http.StatusOK)

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "SUCCESS")
	ctx.Response.SetStatusCode(http.StatusOK)

	if status != data.WithdrawSuccess {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
