package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/oepay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"net/http"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

// oepay支付回调(1：代收  2:代付)
func oepayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	ctx.Response.Header.Set("Content-type", "text/plain")
	switch string(ctx.Method()) {
	case "POST":
	default:
		ctx.Response.SetStatusCode(http.StatusMethodNotAllowed)
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("oepay notify clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	// todo -> oepay
	paramMap, err1 := oepay.Config.ParsePayResult(body)
	if err1 != nil {
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)

	serviceCode := string(ctx.Request.Header.Peek("X-SERVICE-CODE"))
	outSign := string(ctx.Request.Header.Peek("X-SIGN"))
	// 验证签名
	// if serviceCode != oepay.Config.ChannelId {}
	glog.Infof("oppay notice serviceCode: %s", serviceCode)

	orderid := paramMap["orderNo"]
	businessid := service.GetOrderById(orderid, rtype)
	verify := oepay.Config.VerifySing(outSign, paramMap)
	if !verify {
		// 签名验证失败
		PayPointRecord(businessid, "oepay回调", "[oepay] order callback verify sign fail", outSign, rtype)

		glog.Errorf("[oepay] order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}
	switch rtype {
	case 1:
		oepayIncome(ctx, paramMap, businessid)
	case 2:
		oepayWithdraw(ctx, paramMap)
	}
}

// 充值回调
func oepayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["status"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[oepay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "oepay回调", "oepay回调代收验签通过", string(jsonStr), 1)

	if order_status != "1" {
		status = 3
		glog.Errorf("[oepay] order %s callback fail, status: %s", orderid, order_status)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "oepay回调", "oepay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[oepay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "oepay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.OEPAY {
		strmsg := "[oepay] order " + orderid + " is not operable,The order status is " + fmt.Sprint(info.OrderStatus)
		PayPointRecord(businessid, "oepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = time.Now().UnixMilli()
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[oepay] order " + orderid + "Update order error,msg:" + err.Error()
		PayPointRecord(businessid, "oepay回调", strmsg, orderid, 1)
	}

	// 通知服务器
	amount, err := strconv.ParseFloat(order_amount, 64) // 金额(元)
	if err != nil {
		glog.Errorf("oepay order amount parse error: %s", order_amount)
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
		strmsg := "[oepay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "oepay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "success")
		return
	}
	strmsg := "[oepay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "oepay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "success")

	if status == 2 {
		// 订单记录
		OrederRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 提现回调
func oepayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["status"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[oepay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status != "1" {
		glog.Errorf("[oepay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[oepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.OEPAY, status) {
		glog.Warning("[oepay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[oepay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = time.Now().UnixMilli()
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[oepay] order " + orderid + "Update order error,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount, err := strconv.ParseFloat(order_amount, 64) // 金额(元)
	if err != nil {
		glog.Errorf("oepay withdraw order amount parse error: %s", order_amount)
	}

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = int(amount * 100)
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[oepay] order " + orderid + " withdraw callback notify fail,err:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
		fmt.Fprint(ctx, "success")

		// tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
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
