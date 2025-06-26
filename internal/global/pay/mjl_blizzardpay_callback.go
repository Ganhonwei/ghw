package pay

import (
	"encoding/json"
	"fmt"
	mjlblizzardpay "goserver/internal/global/pay/mjlpay/mjlblizzardpay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/valyala/fasthttp"
)

// mjlblizzardpay支付回调(1：代收  2:代付)
func mjlblizzardpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	// body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	postArgs := ctx.PostArgs()

	paramMap := make(map[string]string)

	postArgs.VisitAll(func(key []byte, value []byte) {
		paramMap[string(key)] = string(value)
	})

	// var err1 error
	// if rtype == 1 {
	// 	paramMap, err1 = blizzardpay.Config.ParsePayResult(body)
	// } else {
	// 	paramMap, err1 = blizzardpay.Config.ParsePayResult(body)
	// }
	// if err1 != nil {
	// 	fmt.Fprintf(ctx, "%s", err1)
	// 	return
	// }
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	var key string = ""
	var orderid string = ""
	var businessid string = ""
	switch rtype {
	case 1:
		orderid = paramMap["outTradeNo"]
		key = mjlblizzardpay.Config.Key
	case 2:
		orderid = paramMap["outOrderNo"]
		key = mjlblizzardpay.Config.Key
	}
	sign := mjlblizzardpay.PaySign(paramMap, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "MjlBlizzardPay回调", "[MjlBlizzardPay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[MjlBlizzardPay]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		mjlblizzardpayIncome(ctx, paramMap, businessid)
	case 2:
		mjlblizzardpayWithdraw(ctx, paramMap)
	}
}

// mjlblizzardpay充值回调
func mjlblizzardpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["outTradeNo"]
	order_status := paramMap["payStatus"]
	// order_time := paramMap["timestamp"]
	order_amount := utils.Float64(paramMap["amountTrue"])

	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[MjlBlizzardPay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "MjlBlizzardPay回调", "MjlBlizzardPay回调代收验签通过", string(jsonStr), 1)
	if order_status != "SUCCESS" {
		status = 3
		glog.Errorf("[MjlBlizzardPay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "MjlBlizzardPay回调", "MjlBlizzardPay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[MjlBlizzardPay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "MjlBlizzardPay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJLBLIZZARDPAY {
		strmsg := "[MjlBlizzardPay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "MjlBlizzardPay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SUCCESS")
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
	info.PayTime = time.Now().UnixMilli()
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[MjlBlizzardPay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "MjlBlizzardPay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	// f, err := strconv.ParseFloat(order_amount, 64)
	// if err != nil {
	// 	strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
	// 	PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	// 	glog.Errorf(strmsg)
	// 	// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	// }

	amount := order_amount * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = time.Now().UnixMilli()
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[MjlBlizzardPay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "MjlBlizzardPay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[MjlBlizzardPay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "MjlBlizzardPay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// mjlblizzardpay提现回调
func mjlblizzardpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["outOrderNo"]
	order_status := paramMap["orderStatus"]
	order_time := time.Now().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])

	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[MjlBlizzardPay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		glog.Errorf("[MjlBlizzardPay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[MjlBlizzardPay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJLBLIZZARDPAY, status) {
		glog.Warning("[MjlBlizzardPay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[MjlBlizzardPay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
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
	info.PayTime = order_time
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

	// i := int(f)
	amount := order_amount * 100

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[MjlBlizzardPay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
		fmt.Fprint(ctx, "SUCCESS")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "SUCCESS")

	if status != data.WithdrawSuccess {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}
