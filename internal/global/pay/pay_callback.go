package pay

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	pay1916 "goserver/internal/global/pay/1916pay"
	pay9s "goserver/internal/global/pay/9spay"
	blizzardpay "goserver/internal/global/pay/blizzardpay"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/flypay"
	"goserver/internal/global/pay/kingpay"
	"goserver/internal/global/pay/letspay"
	"goserver/internal/global/pay/mlpay"
	"goserver/internal/global/pay/sailspay"
	"goserver/internal/global/pay/service"
	"goserver/internal/global/pay/tansafe"
	tasks "goserver/internal/global/pay/task"
	"goserver/internal/global/pay/uwinpay"
	"goserver/internal/global/pay/xdpay"
	"goserver/internal/global/pay/xfpay"
	"strconv"
	"strings"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

// ml支付回调(1：代收  2:代付)
func mlpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	ctx.Response.Header.Set("Content-type", "text/plain;charset=UTF-8")
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("clientIP: %s", clientIP)
	//解析
	body := ctx.QueryArgs().QueryString()
	glog.Info("body %s", string(body))
	paramMap := parseParam(body)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	// msg := new(pb.PayOrderUpdate)
	// msg.OutId = paramMap["orderNo"]
	// msg.OrderId = paramMap["partnerOrderNo"]
	// msg.State = data.TradeSuccess
	var key string = ""
	var orderid string = ""
	var businessid string = ""
	switch rtype {
	case 1:
		orderid = paramMap["partnerOrderNo"]
		key = mlpay.Config.Md5Key
	case 2:
		orderid = paramMap["partnerWithdrawNo"]
		key = mlpay.Config.WithDrawMDd5Key
	}
	sign := mlpay.PaySign(paramMap, key)
	upper := data.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != upper {
		// 签名验证失败
		PayPointRecord(businessid, "MLPAY回调", "[mlpay] order callback verify sign fail", orderid, rtype)

		glog.Errorf("[mlpay]order %s withdraw callback verify sign fail", orderid)
		fmt.Fprint(ctx, 1)
		return
	}
	switch rtype {
	case 1:
		mlpayIncome(ctx, paramMap, businessid)
	case 2:
		mlpayWithdraw(ctx, paramMap)
	}
}

// ml充值回调
func mlpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["partnerOrderNo"]
	order_status := paramMap["status"]
	tradeNo := paramMap["orderNo"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[mlpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "MLPAY回调", "MLPAY回调代收验签通过", string(jsonStr), 1)
	if order_status != "1" {
		status = 3
		glog.Errorf("[mlpay] order %s callback fail ", orderid)

	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "MLPAY回调", "MLPAY回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[mlpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.MLPLAY {
		strmsg := "[mlpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:" + err.Error()
		PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, 0)
		return
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[mlpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)
	}
	// 通知服务器
	amount, err := strconv.Atoi(order_amount)
	if err != nil {
		strmsg := "[mlpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[mlpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, 0)
		return
	}
	strmsg := "[mlpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "MLPAY回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, 0)

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// ml提现回调
func mlpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["partnerWithdrawNo"]
	order_status := paramMap["status"]
	msg := paramMap["errorMsg"]
	tradeNo := paramMap["withdrawNo"]
	order_amount := paramMap["amount"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[mlpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		status = data.WithdrawFail
		glog.Errorf("[mlpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[mlpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[mlpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.MLPLAY, status) {
		glog.Warning("[MLPLAY] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[mlpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, 0)
		return
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[mlpay] order " + orderid + "Update order error,msg:" + err.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount, err := strconv.Atoi(order_amount)
	if err != nil {
		strmsg := "[mlpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	order.ChannelId = info.ChannelId
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[mlpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, 0)

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, 0)

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}

// pay1916支付回调(1：代收  2:代付)
func pay1916Notify(ctx *fasthttp.RequestCtx, rtype int) {
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
	var paramMap map[string]interface{}
	var err1 error
	if rtype == 1 {
		paramMap, err1 = pay1916.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = pay1916.Config.ParsePayResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	var key string = ""
	var orderid string = ""
	var businessid string = ""
	switch rtype {
	case 1:
		orderid = paramMap["orderNo"].(string)
		key = pay1916.Config.Md5Key
	case 2:
		orderid = paramMap["orderNo"].(string)
		key = pay1916.Config.WithDrawMDd5Key
	}
	sign := pay1916.PaySign(paramMap, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "PAY1916回调", "[pay1916] order callback verify sign fail", sign, rtype)

		glog.Errorf("[pay1916]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		pay1916Income(ctx, paramMap, businessid)
	case 2:
		pay1916Withdraw(ctx, paramMap)
	}
}

// 1916充值回调
func pay1916Income(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}, businessid string) {
	status := 0
	orderid := paramMap["orderNo"].(string)
	order_status := paramMap["state"].(int)
	order_time := paramMap["timestamp"].(int64)
	order_amount := paramMap["amount"].(int)
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[pay1916] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "PAY1916回调", "PAY1916回调代收验签通过", string(jsonStr), 1)
	if order_status != 1 {
		status = 3
		glog.Errorf("[pay1916] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "PAY1916回调", "PAY1916回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[pay1916] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "PAY1916回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.PAY1916 {
		strmsg := "[pay1916] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "PAY1916回调", strmsg, orderid, 1)

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
	info.OutTradeStatus = strconv.Itoa(order_status)
	info.PayTime = order_time
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[pay1916] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "PAY1916回调", strmsg, orderid, 1)
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
	order.Amount = amount
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[pay1916] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "PAY1916回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[1916pay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "PAY1916回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 1916提现回调
func pay1916Withdraw(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}) {
	status := 0
	orderid := paramMap["orderNo"].(string)
	order_status := paramMap["state"].(int)
	order_time := paramMap["timestamp"].(int64)
	order_amount := paramMap["amount"].(int)
	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[pay1916] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != 1 {
		glog.Errorf("[pay1916] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[pay1916] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.PAY1916, status) {
		glog.Warning("[pay1916] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[pay1916] order %s is not operable", orderid)
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

	info.OutTradeStatus = strconv.Itoa(order_status)
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
	order.Amount = amount
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[pay1916] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

// blizzardpay支付回调(1：代收  2:代付)
func blizzardpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
		key = blizzardpay.Config.Key
	case 2:
		orderid = paramMap["outOrderNo"]
		key = blizzardpay.Config.Key
	}
	sign := blizzardpay.PaySign(paramMap, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "BlizzardPay回调", "[BlizzardPay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[BlizzardPay]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		blizzardpayIncome(ctx, paramMap, businessid)
	case 2:
		blizzardpayWithdraw(ctx, paramMap)
	}
}

// blizzardpay充值回调
func blizzardpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["outTradeNo"]
	order_status := paramMap["payStatus"]
	// order_time := paramMap["timestamp"]
	order_amount := utils.Float64(paramMap["amount"])

	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[BlizzardPay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "BlizzardPay回调", "BlizzardPay回调代收验签通过", string(jsonStr), 1)
	if order_status != "SUCCESS" {
		status = 3
		glog.Errorf("[BlizzardPay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "BlizzardPay回调", "BlizzardPay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[BlizzardPay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "BlizzardPay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.BLIZZARDPY {
		strmsg := "[BlizzardPay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "BlizzardPay回调", strmsg, orderid, 1)

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
		strmsg := "[BlizzardPay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "BlizzardPay回调", strmsg, orderid, 1)
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
		strmsg := "[BlizzardPay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "BlizzardPay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[BlizzardPay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "BlizzardPay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// blizzardpay提现回调
func blizzardpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["outOrderNo"]
	order_status := paramMap["orderStatus"]
	order_time := time.Now().UnixMilli()
	order_amount := utils.Float64(paramMap["amount"])

	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[BlizzardPay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		errMsg := paramMap["errMsg"]
		if strings.Contains(errMsg, "Insufficient") {
			status = data.WithdrawFailTransfer
		} else if strings.Contains(errMsg, "account") ||
			strings.Contains(errMsg, "frozen") {
			status = data.WithdrawFailRefund
		} else {
			status = data.WithdrawFail
		}
		glog.Errorf("[BlizzardPay] order %s callback fail", orderid)
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[BlizzardPay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.BLIZZARDPY, status) {
		glog.Warning("[BlizzardPay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[BlizzardPay] order %s is not operable", orderid)
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
		strmsg := "[BlizzardPay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

// xf支付回调(1：代收  2:代付)
func xfpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
		paramMap, err1 = xfpay.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = xfpay.Config.ParseWithdrawResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	var key string = ""
	var orderid string = ""
	var businessid string = ""
	switch rtype {
	case 1:
		orderid = paramMap["merchantOrderNo"]
		key = xfpay.Config.Md5Key
	case 2:
		orderid = paramMap["merchantOrderNo"]
		key = xfpay.Config.WithDrawMDd5Key
	}
	sign := xfpay.PaySign(paramMap, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "XFPAY回调", "[xfpay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[xfpay]order %s verify sign fail", paramMap["merchantOrderNo"])
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	switch rtype {
	case 1:
		xfpayIncome(ctx, paramMap, businessid)
	case 2:
		xfpayWithdraw(ctx, paramMap)
	}
}

// xf充值回调
func xfpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["merchantOrderNo"]
	order_status := paramMap["status"]
	order_time := paramMap["payTime"]
	order_amount := paramMap["payMoney"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[xfpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "XFPAY回调", "XFPAY回调代收验签通过", string(jsonStr), 1)
	if order_status != "2" {
		status = 3
		glog.Errorf("[xfpay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "XFPAY回调", "XFPAY回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.XFPAY {
		strmsg := "[xfpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	timeStr := order_time                   // 时间字符串
	layout := "Mon Jan 2 15:04:05 MST 2006" // 时间字符串的格式

	t, err := time.Parse(layout, timeStr)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " payment time conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := t.UnixNano() / int64(time.Millisecond)
	info.OutTradeStatus = order_status
	info.PayTime = milliseconds
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[xfpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}

	i := int(f)
	amount := i * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[xfpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[xfpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "XFPAY回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// xf提现回调
func xfpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["merchantOrderNo"]
	order_status := paramMap["status"]
	order_time := paramMap["withdrawalTime"]
	order_amount := paramMap["withdrawalMoney"]
	msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[xfpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "2" {
		glog.Errorf("[xfpay] order %s callback fail msg:%s", orderid, msg)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.XFPAY, status) {
		glog.Warning("[xfpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[xfpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	layout := "Mon Jan 2 15:04:05 MST 2006" // 时间字符串的格式

	t, err := time.Parse(layout, order_time)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " payment time conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := t.UnixNano() / int64(time.Millisecond)

	info.OutTradeStatus = order_status
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[xfpay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[xfpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}

	i := int(f)
	amount := i * 100

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[xfpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s withdraw callback notify fail,err:%s", orderid, err)
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

// king支付回调(1：代收  2:代付)
func kingpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	var orderid string = ""
	var businessid string = ""
	if rtype == 1 {
		orderid = paramMap["outOrderId"]
		paramMap, err1 = kingpay.Config.ParsePayResult(body)
	} else {
		orderid = paramMap["outOrderId"]
		paramMap, err1 = kingpay.Config.ParseWithdrawResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	businessid = service.GetOrderById(orderid, rtype)
	if !kingpay.Config.VerifySing(outSign, kingpay.SignContent(paramMap), crypto.SHA256) {
		// 签名验证失败
		PayPointRecord(businessid, "KingPay回调", "[kingpay] order callback verify sign fail", outSign, rtype)

		glog.Errorf("[kingpay] order %s verify sign fail", orderid)
		res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "1", Message: "verify sign fail"})
		fmt.Fprint(ctx, res)
		return
	}
	switch rtype {
	case 1:
		kingpayIncome(ctx, paramMap, businessid)
	case 2:
		kingpayWithdraw(ctx, paramMap)
	}
}

// king充值回调
func kingpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "0", Message: "SUCCESS"})
	status := 0
	orderid := paramMap["outOrderId"]
	order_status := paramMap["status"]
	order_time := paramMap["payTime"]
	tradeNo := paramMap["orderId"]
	order_amount := paramMap["amount"]
	msg := paramMap["statusDesc"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[kingpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "KingPay回调", "KingPay回调代收验签通过", string(jsonStr), 1)
	if order_status != "4" {
		status = 3
		glog.Errorf("[kingpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	if err != nil {
		glog.Errorf("[kingpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	PayPointRecord(businessid, "KingPay回调", "KingPay回调通过商户ID查询支付订单", orderid, 1)
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.KINGPAY {
		strmsg := "[kingpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "KingPay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, res)
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[kingpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := int64(timestamp) // 示例的毫秒级时间戳
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[kingpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "KingPay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[kingpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "KingPay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[kingpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}
	i := int(f)
	amount := i * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[kingpay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "KingPay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, res)
		return
	}
	strmsg := "[kingpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "KingPay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, res)

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// king提现回调
func kingpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "0", Message: "SUCCESS"})
	status := 0
	orderid := paramMap["outOrderId"]
	order_status := paramMap["status"]
	order_time := paramMap["transferTime"]
	tradeNo := paramMap["orderId"]
	order_amount := paramMap["withdrawAmount"]
	msg := paramMap["statusDesc"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[kingpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "5" {
		glog.Errorf("[kingpay] order %s callback fail msg:%s", orderid, msg)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[kingpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.XFPAY, status) {
		glog.Warning("[xfpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[kingpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, res)
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[kingpay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := int64(timestamp) // 示例的毫秒级时间戳

	str := string(jsonStr)

	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[kingpay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[kingpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	i := int(f)
	amount := i * 100

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[kingpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, res)

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, res)

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}

// sailspay支付回调(1：代收  2:代付)
func sailspayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	var orderid string = ""
	var businessid string = ""
	if rtype == 1 {
		orderid = paramMap["orderId"]
		paramMap, err1 = sailspay.Config.ParsePayResult(body)
	} else {
		orderid = paramMap["orderId"]
		paramMap, err1 = sailspay.Config.ParseWithdrawResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	key := sailspay.Config.Md5Key
	orderId := paramMap["orderId"]
	payTime := paramMap["payTime"]
	sign := sailspay.PaySign(orderId, payTime, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		glog.Errorf("[sailspay]order %s verify sign fail", paramMap["orderId"])
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	switch rtype {
	case 1:
		sailspayIncome(ctx, paramMap, businessid)
	case 2:
		sailspayWithdraw(ctx, paramMap)
	}
}

// sailspay充值回调
func sailspayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	order_time := paramMap["payTime"]
	tradeNo := paramMap["payOrderId"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[sailspay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "sailspay回调", "sailspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "1" {
		status = 3
		glog.Errorf("[sailspay] order %s callback fail", orderid)
	} else {
		status = 2
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[sailspay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := int64(timestamp) // 示例的毫秒级时间戳

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "sailspay回调", "sailspay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		glog.Errorf("[sailspay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.SAILSPAY {
		strmsg := "[sailspay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "sailspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[sailspay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[sailspay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "sailspay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	amount, err := strconv.Atoi(order_amount)
	if err != nil {
		strmsg := "[sailspay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "sailspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[sailspay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "sailspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[sailspay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "sailspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// sailspay提现回调
func sailspayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	order_time := paramMap["payOrderId"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[sailspay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		status = data.WithdrawFail
		glog.Errorf("[sailspay] order %s callback fail", orderid)

	} else {
		status = data.WithdrawSuccess
	}
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[sailspay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.SAILSPAY, status) {
		glog.Warning("[sailspay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[sailspay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[sailspay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := int64(timestamp) // 示例的毫秒级时间戳

	str := string(jsonStr)

	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[sailspay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount, err := strconv.Atoi(order_amount)
	if err != nil {
		strmsg := "[sailspay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[sailspay] order " + orderid + " pay callback notify fail,err:" + err.Error()
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

// flypay支付回调(1：代收  2:代付)
func flypayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	var orderid string = ""
	var businessid string = ""
	if rtype == 1 {
		paramMap, err1 = flypay.Config.ParsePayResult(body)
		orderid = paramMap["mer_order_no"]
	} else {
		paramMap, err1 = flypay.Config.ParseWithdrawResult(body)
		orderid = paramMap["mer_order_no"]
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	businessid = service.GetOrderById(orderid, rtype)
	if rtype == 1 {
		// 代收
		key := flypay.Config.Md5Key
		machid := flypay.Config.MachId
		sign := flypay.PaySign(machid, orderid, key)
		if outSign != sign {
			// 签名验证失败
			PayPointRecord(businessid, "flypay回调", "[flypay] order callback verify sign fail", orderid, rtype)

			glog.Errorf("[flypay] order %s verify sign fail,sign:%s", orderid, sign)
			fmt.Fprint(ctx, "Sign Fail")
			return
		}
		flypayIncome(ctx, paramMap, businessid)
	} else if rtype == 2 {
		// 代付
		strbyte, _ := json.Marshal(paramMap)
		fmt.Print(string(strbyte))

		if !flypay.Config.VerifySing(outSign, strbyte) {
			// 签名验证失败
			glog.Errorf("[flypay] order %s verify sign fail", orderid)
			res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "1", Message: "verify sign fail"})
			fmt.Fprint(ctx, res)
			return
		}
		flypayWithdraw(ctx, paramMap)
	}
}

// flypay充值回调
func flypayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	orderid := paramMap["mer_order_no"]
	order_status := paramMap["order_status"]
	msg := paramMap["msg"]
	order_time := paramMap["order_time"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["pay_amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[flypay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "flypay回调", "flypay回调代收验签通过", string(jsonStr), 1)
	if paramMap["code"] == "0" {
		glog.Errorf("[flypay] Interface access failure,msg:%s", msg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	if order_status != "4" {
		status = 3
		glog.Errorf("[flypay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[flypay]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := timestamp * 1000 // 转换为毫秒级时间戳

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "flypay回调", "flypay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[flypay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.FLYPAY {
		strmsg := "[flypay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:" + err.Error()
		PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	info.RefuseReason = msg
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[flypay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[flypay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[flypay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}
	amount := int(f)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[flypay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[flypay] order %s pay callback notify fail,err:%s", orderid, err1.Error())
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[flypay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "flypay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// flypay提现回调
func flypayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	orderid := paramMap["mer_order_no"]
	order_status := paramMap["order_status"]
	msg := paramMap["msg"]
	order_time := paramMap["order_time"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["order_amount"]
	status := 0
	if paramMap["code"] == "0" {
		glog.Errorf("[flypay] Interface access failure,msg:%s", msg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[flypay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "4" {
		glog.Errorf("[flypay] order %s callback fail msg:%s", orderid, msg)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[flypay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[flypay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.FLYPAY, status) {
		glog.Warning("[flypay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[flypay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		strmsg := "[flypay] order " + orderid + " payment time conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	milliseconds := timestamp * 1000 // 转换为毫秒级时间戳

	str := string(jsonStr)

	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[flypay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	amount, err := strconv.Atoi(order_amount)
	if err != nil {
		glog.Errorf("[flypay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[flypay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

// xdpay支付回调(1：代收  2:代付)
func xdpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := xdpay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	key := xdpay.Config.Md5Key
	sign := xdpay.PaySign(paramMap, key)
	var orderid string = ""
	var businessid string = ""
	orderid = paramMap["orderId"]
	businessid = service.GetOrderById(orderid, rtype)
	// upper := pay.ToUpper(sign)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "xdpay回调", "[xdpay] order callback verify sign fail", orderid, rtype)

		glog.Errorf("[xdpay] order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "success")
		return
	}
	switch rtype {
	case 1:
		xdpayIncome(ctx, paramMap, businessid)
	case 2:
		xdpayWithdraw(ctx, paramMap)
	}
}

// xdpay充值回调
func xdpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	msg := paramMap["remark"]
	tradeNo := paramMap["platOrderId"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[xdpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "xdpay回调", "xdpay回调代收验签通过", string(jsonStr), 1)
	if order_status != "1" {
		status = 3
		glog.Errorf("[xdpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "xdpay回调", "xdpay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[xdpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.XDPAY {
		strmsg := "[xdpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:" + err.Error()
		PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "success")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	info.RefuseReason = msg
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[xdpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[xdpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	amount := int(f)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[xdpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "success")
		return
	}
	strmsg := "[xdpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "xdpay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "success")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// xdpay提现回调
func xdpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	orderid := paramMap["orderId"]
	order_status := paramMap["status"]
	msg := paramMap["remark"]
	tradeNo := paramMap["platOrderId"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[xdpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "1" {
		status = data.WithdrawFail
		glog.Errorf("[xdpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = data.WithdrawSuccess
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[xdpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xdpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.XDPAY, status) {
		glog.Warning("[xdpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[xdpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "success")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	info.RefuseReason = msg
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[xdpay] order " + orderid + "Update order error,msg:" + err1.Error()
		glog.Errorf(strmsg)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[xdpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	amount := int(f)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[xdpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		// glog.Errorf("[xdpay] order %s withdraw callback notify fail,err:%s", orderid, err1.Error())
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
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}

// uwinpay支付回调(1：代收  2:代付)
func uwinpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := uwinpay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	content := uwinpay.SignContent(paramMap)
	sign := uwinpay.Config.VerifySing(outSign, content, crypto.SHA256)
	// upper := pay.ToUpper(sign)
	var orderid string = ""
	var businessid string = ""
	orderid = paramMap["merchant_order_no"]
	businessid = service.GetOrderById(orderid, rtype)
	if !sign {
		// 签名验证失败
		PayPointRecord(businessid, "uwinpay回调", "[uwinpay] order callback verify sign fail", orderid, rtype)

		glog.Errorf("[uwinpay] order %s verify sign fail", paramMap["merchant_order_no"])
		fmt.Fprint(ctx, "ok")
		return
	}
	if paramMap["merchant_code"] != uwinpay.Config.Merchant {
		// 签名验证失败
		PayPointRecord(businessid, "uwinpay回调", "[uwinpay] order merchant_code exception", orderid, rtype)

		glog.Errorf("[uwinpay] order %s merchant_code exception", paramMap["merchant_order_no"])
		fmt.Fprint(ctx, "ok")
		return
	}
	switch rtype {
	case 1:
		uwinpayIncome(ctx, paramMap, businessid)
	case 2:
		uwinpayWithdraw(ctx, paramMap)
	}
}

// uwinpay充值回调
func uwinpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	orderid := paramMap["merchant_order_no"]
	order_status := paramMap["status"]
	msg := paramMap["error_message"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[uwinpay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "uwinpay回调", "uwinpay回调代收验签通过", string(jsonStr), 1)
	if order_status != "2" {
		status = 3
		glog.Errorf("[uwinpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "uwinpay回调", "uwinPAY回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[uwinpay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.UWINPAY {
		strmsg := "[uwinpay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "ok")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	info.RefuseReason = msg
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[uwinpay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[uwinpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[uwinpay]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}
	amount := int(f)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[uwinpay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "ok")
		return
	}
	strmsg := "[uwinpay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "uwinpay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "ok")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// uwinpay提现回调
func uwinpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	orderid := paramMap["merchant_order_no"]
	order_status := paramMap["status"]
	msg := paramMap["error_message"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[uwinpay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "2" {
		status = data.WithdrawFail
		glog.Errorf("[uwinpay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = data.WithdrawSuccess
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[uwinpay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.UWINPAY, status) {
		glog.Warning("[uwinpay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Infof("[uwinpay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "ok")
		return
	}
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[uwinpay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[uwinpay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		glog.Errorf(strmsg)
	}
	amount := int(f)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = int64(milliseconds)
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[uwinpay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

// tansafe支付回调(1：代收  2:代付)
func tansafeNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
		paramMap, err1 = tansafe.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = tansafe.Config.ParseWithdrawResult(body)
	}
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	var orderid string = ""
	var businessid string = ""
	orderid = paramMap["mchOrderNo"]
	stringByte, _ := base64.StdEncoding.DecodeString(outSign)
	businessid = service.GetOrderById(orderid, rtype)
	if !tansafe.Config.Verify(tansafe.SignContent(paramMap), stringByte, crypto.SHA1) {
		// 签名验证失败
		PayPointRecord(businessid, "tansafe回调", "[tansafe] order callback verify sign fail", orderid, rtype)

		glog.Errorf("[tansafe] order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	switch rtype {
	case 1:
		tansafeIncome(ctx, paramMap, businessid)
	case 2:
		tansafeWithdraw(ctx, paramMap)
	}
}

// tansafe充值回调
func tansafeIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["mchOrderNo"]
	order_status := paramMap["orderStatus"]
	reqstatus := paramMap["status"]
	order_time := paramMap["timeEnd"]
	tradeNo := paramMap["orderStatus"]
	order_amount := paramMap["amount"]
	msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[tansafe] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "tansafe回调", "tansafe回调代收验签通过", string(jsonStr), 1)
	if reqstatus != "200" && order_status == "SUCCESS" {
		status = 3
		glog.Errorf("[tansafe] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}
	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "tansafe回调", "tansafe回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[tansafe] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		strmsg := "[tansafe] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[tansafe]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}

	milliseconds := int64(timestamp) * 1000 // 示例的毫秒级时间戳
	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[tansafe] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[tansafe] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	i := int(f)
	amount := i * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = amount
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[tansafe] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[tansafe] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "tansafe回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	// 成功需要记录日志
	if status == 2 {
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// tansafe提现回调
func tansafeWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "0", Message: "SUCCESS"})
	status := 0
	orderid := paramMap["mchOrderNo"]
	order_status := paramMap["orderStatus"]
	reqstatus := paramMap["status"]
	order_time := paramMap["timeEnd"]
	tradeNo := paramMap["orderStatus"]
	order_amount := paramMap["amount"]
	msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[tansafe] Convert to json exception,msg:%s", err.Error())
	}
	if reqstatus != "200" && order_status == "SUCCESS" {
		glog.Errorf("[tansafe] order %s callback fail msg:%s", orderid, msg)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[tansafe] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.TANSAFEPAY, status) {
		glog.Warning("[tansafe] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[tansafe] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, res)
		return
	}
	timestamp, err := strconv.Atoi(order_time)
	if err != nil {
		glog.Errorf("[tansafe]order %s payment time conversion failed,msg:%s", orderid, err.Error())
	}
	milliseconds := int64(timestamp) * 1000 // 示例的毫秒级时间戳

	str := string(jsonStr)

	info.OutTradeStatus = order_status
	info.OutTradeNo = tradeNo
	info.PayTime = milliseconds
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	info.RefuseReason = msg
	// err1 := service.UpdateWithDrawOrder(info)
	// if err1 != nil {
	// 	strmsg := "[tansafe] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		glog.Errorf("[tansafe]order %s amount conversion failed,msg:%s", orderid, err.Error())
	}

	i := int(f)
	amount := i * 100

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount * 100
	order.Timestamp = milliseconds
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[tansafe] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, res)

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, res)

	if status != 2 {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}

// letsPay支付回调(1：代收  2:代付)
func letsPayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	paramMap, err1 := letspay.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	// glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	delete(paramMap, "msg")
	var businessid string = ""

	orderid := paramMap["orderNo"]
	key := letspay.Config.Md5Key
	sign := letspay.PaySign(paramMap, key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "letspay回调", "[letspay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[letspay]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		letsPayIncome(ctx, paramMap, businessid)
	case 2:
		letsPayWithdraw(ctx, paramMap)
	}
}

// letsPay充值回调
func letsPayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	status := 0
	orderid := paramMap["orderNo"]
	order_status := paramMap["status"]
	order_time := utils.Int64(paramMap["paySuccTime"])
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[letspay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "letspay回调", "letspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "2" {
		status = 3
		glog.Errorf("[letspay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "letspay回调", "letspay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[letspay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.LETSPAY {
		strmsg := "[letspay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[letspay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	}

	amount := order_amount * 100
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = order_time
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[letspay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[letspay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "letspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// letspay提现回调
func letsPayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	status := 0
	orderid := paramMap["mchTransNo"]
	order_status := paramMap["status"]
	order_time := utils.Int64(paramMap["transSuccTime"])
	order_amount := utils.Float64(paramMap["amount"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[letspay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "2" {
		glog.Errorf("[letspay] order %s callback fail", orderid)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[letspay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.LETSPAY, status) {
		glog.Warning("[letspay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[letspay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = order_time
	info.OrderStatus = int32(status)
	info.CallBackMsg = str
	err1 := service.UpdateWithDrawOrder(info)
	if err1 != nil {
		strmsg := "[letspay] order " + orderid + "Update order error,msg:" + err1.Error()
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
		strmsg := "[letspay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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

// pay9s支付回调(1：代收  2:代付)
func pay9sNotify(ctx *fasthttp.RequestCtx, rtype int) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "fail")
		}
	}()
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		fmt.Fprintf(ctx, "%s", "fail")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	paramMap, err1 := pay9s.Config.ParsePayResult(body)
	if err1 != nil {
		fmt.Fprintf(ctx, "%s", err1)
		return
	}
	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	orderid := paramMap["merchantOrderNo"].(string)
	// 验证签名
	delete(paramMap, "sign")

	var businessid string = ""

	sign := pay9s.PaySign(paramMap, pay9s.Config.Md5Key)
	// upper := pay.ToUpper(sign)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "pay9s回调", "[pay9s] order callback verify sign fail", sign, rtype)

		glog.Errorf("[pay9s]order %s verify sign fail", paramMap["merchantOrderNo"])
		fmt.Fprint(ctx, "fail")
		return
	}
	switch rtype {
	case 1:
		pay9sIncome(ctx, paramMap, businessid)
	case 2:
		pay9sWithdraw(ctx, paramMap)
	}
}

// 9spay充值回调
func pay9sIncome(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}, businessid string) {
	status := 0
	orderid := paramMap["merchantOrderNo"].(string)
	order_status := paramMap["status"].(string)
	order_time := paramMap["createOrderTime"].(string)
	order_amount := utils.Float64(paramMap["amount"].(string))
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[pay9s] Convert to json exception,msg:%s", err.Error())
	}

	PayPointRecord(businessid, "pay9s回调", "pay9s回调代收验签通过", string(jsonStr), 1)
	if order_status != "PAID" {
		status = 3
		glog.Errorf("[pay9s] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "pay9s回调", "pay9s回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[pay9s] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "pay9s回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.PAY9S {
		strmsg := "[pay9s] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "pay9s回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = utils.Int64(order_time)
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[pay9s] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "pay9s回调", strmsg, orderid, 1)
	}

	amount := int(order_amount * 100)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.BusinessID = businessid
	order.Status = status
	order.Amount = int(amount)
	order.Timestamp = utils.Int64(order_time)
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		strmsg := "[pay9s] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "pay9s回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "success")
		return
	}
	strmsg := "[pay9s] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "pay9s回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "success")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// 9s提现回调
func pay9sWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}) {
	status := 0
	orderid := paramMap["merchantOrderNo"].(string)
	order_status := paramMap["status"].(string)
	order_time := paramMap["createOrderTime"].(string)
	order_amount := utils.Float64(paramMap["amount"].(string))
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[pay9s] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "PAID" {
		glog.Errorf("[pay9s] order %s callback fail, status：%s", orderid, order_status)
		status = data.WithdrawFail
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[pay9s] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.PAY9S, status) {
		glog.Warning("[pay9s] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "success")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[pay9s] order %s is not operable", orderid)
		fmt.Fprint(ctx, "success")
		return
	}

	info.OutTradeStatus = order_status
	info.PayTime = utils.Int64(order_time)
	info.OrderStatus = int32(status)
	info.CallBackMsg = str

	amount := int(order_amount * 100)

	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = utils.Int64(order_time)
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[pay9s] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "success")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "success")

	if status != data.WithdrawSuccess {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}
}

func parseParam(body []byte) map[string]string {
	param := make(map[string]string)
	str := string(body)
	strs := strings.Split(str, "&")
	for _, v := range strs {
		value := strings.Split(v, "=")

		param[value[0]] = value[1]
	}
	return param
}

func IsRepeatPay(order *entity.WithdrawOrder, channelId uint32, status int) bool {
	if order.ChannelId != channelId {
		// 重复付款需要记录
		req := &data.RepeatWithdrawReq{
			OrderID:   order.MerOrderID,
			ChannelID: channelId,
			Status:    status,
			Msg:       "重复付款",
		}
		service.RepeatWithdrawNotify(req)
		// order.OrderStatus = int32(status)

		if status == data.WithdrawSuccess {
			// 成功需要记录日志
			tasks.CreatePayLogSchedule(order.OrderID, order.Userid, 2, channelId, int64(order.Amount))
		}
		return true
	}
	return false
}
