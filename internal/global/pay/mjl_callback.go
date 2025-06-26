package pay

import (
	"crypto"
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/mjlpay/jypay"
	"goserver/internal/global/pay/mjlpay/mmpay"
	"goserver/internal/global/pay/mjlpay/pay99"
	"goserver/internal/global/pay/mjlpay/shpay"
	"goserver/internal/global/pay/mjlpay/transafepay"
	"goserver/internal/global/pay/mjlpay/unipay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strings"

	"github.com/valyala/fasthttp"
)

func uniPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := unipay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")

	sign := service.PaySignMd5String(paramMap, unipay.Config.Appid)
	if sign != outSign {
		glog.Errorf("[twpay]order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		uniPayIncome(ctx, paramMap)
	case 2:
		uniPayWithdraw(ctx, paramMap)
	}
}

// uniPay充值回调
func uniPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_id"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[twpay] Convert to json exception,msg:%s", err.Error())
	}
	// PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "2" {
		status = 3
		glog.Errorf("[twpay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[twpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_UNI {
		strmsg := "[twpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[twpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// unipay提现回调
func uniPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_id"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[twpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "2" {
		glog.Errorf("[twpay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[tw] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_UNI, status) {
		glog.Warning("[tw] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[twpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[twpay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[twpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

func jyPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := jypay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")

	sign := jypay.HmacRSASign(jypay.SignContent(paramMap), jypay.Config.HamcRSA, crypto.SHA256)
	if sign != outSign {
		glog.Errorf("[jypay]order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		jyPayIncome(ctx, paramMap)
	case 2:
		jyPayWithdraw(ctx, paramMap)
	}
}

// jyPay充值回调
func jyPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["merOrderNo"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["payAmount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[jypay] Convert to json exception,msg:%s", err.Error())
	}
	// PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "5" {
		status = 3
		glog.Errorf("[jypay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[jypay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_JY {
		strmsg := "[jypay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[jypay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// jypay提现回调
func jyPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["merOrderNo"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["orderAmount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[jypay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "7" {
		glog.Errorf("[jypay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[jypay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_JY, status) {
		glog.Warning("[jypay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[jypay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[jypay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[jypay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

func _99PayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	// 从请求头中获取 OC-Signature
	ocSignature := string(ctx.Request.Header.Peek("OC-Signature"))
	glog.Infof("OC-Signature: %s", ocSignature)

	//解析
	body := ctx.PostBody()

	paramMap, err1 := pay99.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	sig := pay99.ComputeHMAC(paramMap, pay99.Config.HamcRSA)
	if sig != ocSignature {
		glog.Errorf("[99pay]merchant order %s, order %s verify sign fail", paramMap["merchant_order_id"], paramMap["order_uid"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	// outSign := paramMap["sign"]
	// 验证签名
	// delete(paramMap, "sign")

	// sign := pay99.HmacRSASign(pay99.SignContent(paramMap), pay99.Config.HamcRSA, crypto.SHA256)
	// if sign != outSign {
	// 	glog.Errorf("[jypay]order %s verify sign fail", paramMap["order_id"])
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }
	switch rtype {
	case 1:
		_99PayIncome(ctx, paramMap)
	case 2:
		_99PayWithdraw(ctx, paramMap)
	}
}

// 99Pay充值回调
func _99PayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["merchant_order_id"]   //商户订单（我们自己的订单号）
	order_status := paramMap["payment_status"] //订单状态
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"]) //订单金额
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[99pay] Convert to json exception,msg:%s", err.Error())
	}
	// PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "completed" {
		status = 3
		glog.Errorf("[99pay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[99pay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_99 {
		strmsg := "[99pay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[99pay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 99pay提现回调
func _99PayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["merchant_order_id"]
	order_status := paramMap["payment_status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[99pay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "completed" {
		glog.Errorf("[99pay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[99pay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_99, status) {
		glog.Warning("[99pay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[99pay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[99pay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[99pay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

func transafePayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := transafepay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	// 签名
	if !transafepay.Config.VerifySing(outSign, transafepay.SignContent(paramMap), crypto.SHA1) {
		// 签名验证失败
		glog.Errorf("[transafepay]order %s verify sign fail", paramMap["mchOrderNo"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		transafePayIncome(ctx, paramMap)
	case 2:
		transafePayWithdraw(ctx, paramMap)
	}
}

// transafePay充值回调
func transafePayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["mchOrderNo"]
	order_status := paramMap["orderStatus"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[transafepay] Convert to json exception,msg:%s", err.Error())
	}
	// PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "SUCCESS" {
		status = 3
		glog.Errorf("[transafepay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[transafepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_Transafe {
		strmsg := "[transafepay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[transafepay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// transafepay提现回调
func transafePayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["mchOrderNo"]
	order_status := paramMap["orderStatus"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[transafepay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "SUCCESS" {
		glog.Errorf("[transafepay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[transafepay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_Transafe, status) {
		glog.Warning("[transafepay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[transafepay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[transafepay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[transafepay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

func mmPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := mmpay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")

	sign := service.PaySignMd5StringNoKey(paramMap, mmpay.Config.Md5Key)
	if sign != outSign {
		glog.Errorf("[mmpay]order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		mmPayIncome(ctx, paramMap)
	case 2:
		mmPayWithdraw(ctx, paramMap)
	}
}

// mmPay充值回调
func mmPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_no"]
	order_status := paramMap["status"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["order_reality_amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[mmpay] Convert to json exception,msg:%s", err.Error())
	}
	// PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "success" {
		status = 3
		glog.Errorf("[mmpay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[mmpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_MM {
		strmsg := "[mmpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[mmpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// mmpay提现回调
func mmPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["order_no"]
	order_status := paramMap["result"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["order_amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[mmpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "success" {
		glog.Errorf("[mmpay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[mmpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_MM, status) {
		glog.Warning("[mmpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[mmpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[mmpay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[mmpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

func shPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	//解析
	body := ctx.PostBody()

	paramMap, err1 := shpay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	delete(paramMap, "paymentTransNo")
	delete(paramMap, "reference")
	delete(paramMap, "extInfo")

	sign := service.PaySignMd5StringNoKey(paramMap, shpay.Config.Md5Key)
	sign = strings.ToUpper(sign)
	if sign != outSign {
		glog.Errorf("[shpay]order %s verify sign fail", paramMap["order_id"])
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		shPayIncome(ctx, paramMap)
	case 2:
		shPayWithdraw(ctx, paramMap)
	}
}

// shPay充值回调
func shPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["outTradeNo"]
	order_status := paramMap["transStatus"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["transAmt"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[shpay] Convert to json exception,msg:%s", err.Error())
	}

	if order_status == "FAIL" {
		status = 3
		glog.Errorf("[shpay] order %s callback fail", orderid)
	} else if order_status == "SUCCESS" {
		status = 2
	} else {
		return
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		strmsg := "[shpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MJL_SHPAY {
		strmsg := "[shpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)

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
		strmsg := "[shpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	// strmsg := "[twpay] order " + orderid + " pay callback notify success"
	// PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// shpay提现回调
func shPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["outTradeNo"]
	order_status := paramMap["transStatus"]
	order_time := utils.BsonNow().UnixMilli()
	order_amount := utils.Float64(paramMap["transAmt"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[shpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status == "FAIL" {
		glog.Errorf("[shpay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else if order_status == "SUCCESS" {
		status = data.WithdrawSuccess
	} else {
		return
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[shpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MJL_SHPAY, status) {
		glog.Warning("[shpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[shpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[shpay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[shpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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
