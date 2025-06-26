package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/kkpluspay"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"net/url"
	"strconv"
	"strings"

	tasks "goserver/internal/global/pay/task"

	"github.com/valyala/fasthttp"
)

// kkpluspay支付回调(1：代收  2:代付)
func kkpluspayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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

	strbody := strings.ReplaceAll(string(body), "+", "%20")
	values, err := url.ParseQuery(strbody)
	if err != nil {
		fmt.Fprintf(ctx, "%s", err)
		return
	}

	paramMap := kkpluspay.Config.ValuesToMap(values)

	glog.Debugf("body %v", paramMap)
	outSign := paramMap["sign"]
	// 验证签名
	delete(paramMap, "sign")
	// content := kkpluspay.SignContent(paramMap)

	// sign := kkpluspay.Config.VerifySing(outSign, content, crypto.SHA256)
	sign := kkpluspay.PaySign(paramMap, kkpluspay.Config.MD5Key)
	// upper := pay.ToUpper(sign)
	var orderid string = ""
	var businessid string = ""
	orderid = paramMap["mer_order_no"]
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "kkpluspay回调", "[kkpluspay] order callback verify sign fail", orderid, rtype)

		glog.Errorf("[kkpluspay] order %s verify sign fail", paramMap["merchant_order_no"])
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	if paramMap["mer_no"] != kkpluspay.Config.Merchant {
		// 签名验证失败
		PayPointRecord(businessid, "kkpluspay回调", "[kkpluspay] order merchant_code exception", orderid, rtype)

		glog.Errorf("[kkpluspay] order %s merchant_code exception", paramMap["merchant_order_no"])
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	switch rtype {
	case 1:
		kkpluspayIncome(ctx, paramMap, businessid)
	case 2:
		kkpluspayWithdraw(ctx, paramMap)
	}
}

// kkpluspay充值回调
func kkpluspayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, businessid string) {
	orderid := paramMap["mer_order_no"]
	order_status := paramMap["status"]
	msg := paramMap["err_msg"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["pay_amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[kkpluspay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "kkpluspay回调", "kkpluspay回调代收验签通过", string(jsonStr), 1)
	if order_status != "SUCCESS" {
		status = 3
		glog.Errorf("[kkpluspay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = 2
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "kkpluspay回调", "kkpluspay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[kkpluspay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.KKPLUSPAY {
		strmsg := "[kkpluspay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus)
		PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)

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
		strmsg := "[kkpluspay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)
	}
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[kkpluspay] order " + orderid + " amount conversion failed,msg:" + err.Error()
		PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[kkpluspay]order %s amount conversion failed,msg:%s", orderid, err.Error())
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
		strmsg := "[kkpluspay] order " + orderid + " pay callback notify fail,err:" + err2.Error()
		PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}
	strmsg := "[kkpluspay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "kkpluspay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SUCCESS")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// kkpluspay提现回调
func kkpluspayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
	orderid := paramMap["mer_order_no"]
	order_status := paramMap["status"]
	msg := paramMap["err_msg"]
	tradeNo := paramMap["order_no"]
	order_amount := paramMap["order_amount"]
	status := 0
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[kkpluspay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "SUCCESS" {
		status = data.WithdrawFail
		glog.Errorf("[kkpluspay] order %s callback fail msg:%s", orderid, msg)
	} else {
		status = data.WithdrawSuccess
	}
	milliseconds := utils.BsonNow().UnixNano() / 1e6

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[kkpluspay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.KKPLUSPAY, status) {
		glog.Warning("[kkpluspay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SUCCESS")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Infof("[kkpluspay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SUCCESS")
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
	// 	strmsg := "[kkpluspay] order " + orderid + "Update order error,msg:" + err1.Error()
	// 	glog.Errorf(strmsg)
	// }
	// 通知服务器
	f, err := strconv.ParseFloat(order_amount, 64)
	if err != nil {
		strmsg := "[kkpluspay] order " + orderid + " amount conversion failed,msg:" + err.Error()
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
		strmsg := "[kkpluspay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
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
