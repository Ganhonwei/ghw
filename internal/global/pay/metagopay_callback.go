package pay

import (
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/metagopay"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// metagopay支付回调(1：代收  2:代付)
func metagopayNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	glog.Infof("clientIP: %s", clientIP)
	//解析
	contentType := ctx.Request.Header.ContentType()
	fmt.Println("contentType:", string(contentType))
	var body []byte
	// "multipart/form-data; boundary=--------------------------442509571251475036204101"
	// "application/x-www-form-urlencoded; charset=UTF-8"
	// "orgNo=8241200147&custId=24122000000142&custOrderNo=10005&prdOrdNo=202412240827192311767KZS1EbNA&ordStatus=01&ordAmt=15000&payAmt=15000&ordTime=20241224105719&utr=None&version=2.1&sign=76FAFA389162C31DEDB30EDBC3EB3D6D"
	if strings.HasPrefix(string(contentType), "multipart/form-data") {
		f, err := ctx.MultipartForm()
		if err != nil {
			fmt.Fprintf(ctx, "%s", err)
			return
		}
		var params = make(map[string]interface{})
		for k, vs := range f.Value {
			if len(vs) > 0 {
				params[k] = vs[len(vs)-1]
			}
		}
		body, _ = json.Marshal(params)
	} else if strings.HasPrefix(string(contentType), "application/x-www-form-urlencoded") {
		arg := ctx.Request.PostArgs()
		var params = make(map[string]interface{})
		arg.VisitAll(func(key, value []byte) {
			params[string(key)] = string(value)
		})
		body, _ = json.Marshal(params)
	} else {
		body = ctx.PostBody()
	}

	var paramMap map[string]interface{}
	var err1 error
	if rtype == 1 {
		paramMap, err1 = metagopay.Config.ParsePayResult(body)
	} else {
		paramMap, err1 = metagopay.Config.ParseWithdrawResult(body)
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
		orderid = paramMap["custOrderNo"].(string)
		key = metagopay.Config.Md5Key
	case 2:
		orderid = paramMap["custOrderNo"].(string)
		key = metagopay.Config.Md5Key
	}
	sign := metagopay.PaySign(paramMap, key)
	businessid = service.GetOrderById(orderid, rtype)
	if outSign != sign {
		// 签名验证失败
		PayPointRecord(businessid, "metagopay回调", "[metagopay] order callback verify sign fail", sign, rtype)

		glog.Errorf("[metagopay]order %s verify sign fail", orderid)
		fmt.Fprint(ctx, "failure")
		return
	}
	switch rtype {
	case 1:
		metagopayIncome(ctx, paramMap, businessid)
	case 2:
		metagopayWithdraw(ctx, paramMap)
	}
}

// metagopay充值回调
func metagopayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}, businessid string) {
	status := 0
	orderid := paramMap["custOrderNo"].(string)
	order_status := paramMap["ordStatus"].(string)
	order_time := paramMap["ordTime"].(string)
	order_amount := fmt.Sprint(paramMap["payAmt"])
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[metagopay] Convert to json exception,msg:%s", err.Error())
	}
	PayPointRecord(businessid, "metagopay回调", "metagopay回调代收验签通过", string(jsonStr), 1)
	if order_status != "01" {
		status = 3
		glog.Errorf("[metagopay] order %s callback fail", orderid)
	} else {
		status = 2
	}

	str := string(jsonStr)
	info, err := service.GetPayOrder(orderid)
	PayPointRecord(businessid, "metagopay回调", "metagopay回调通过商户ID查询支付订单", orderid, 1)
	if err != nil {
		strmsg := "[metagopay] order " + orderid + " query exception,fail msg:" + err.Error()
		PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s query exception,fail msg:%s", orderid, err.Error())
	}
	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != service.METAGOPAY {
		strmsg := "[metagopay] order " + orderid + " is not operable,The order status is " + string(info.OrderStatus) + ",fail msg:"
		PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)
		// glog.Errorf("[xfpay] order %s is not operable,fail msg:%s", orderid, err.Error())
		fmt.Fprint(ctx, "SC000000")
		return
	}
	timeStr := order_time      // 时间字符串
	layout := "20060102150405" // 时间字符串的格式
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		strmsg := "[metagopay] order " + orderid + " payment time conversion failed, time:" + timeStr + ", msg:" + err.Error()
		PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)
		glog.Errorf(strmsg)
	}
	info.OutTradeStatus = order_status
	info.PayTime = t.UnixMilli()
	info.CallBackMsg = str
	info.OrderStatus = int32(status)
	err1 := service.UpdateOrder(info)
	if err1 != nil {
		strmsg := "[metagopay] order " + orderid + "Update order error,msg:" + err1.Error()
		PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)
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
	order.Timestamp = t.UnixMilli()
	_, err2 := service.PayNotify(order, 1)
	if err2 != nil {
		// glog.Errorf("[xfpay] order %s pay callback notify fail,err:%s", orderid, err)
		strmsg := "[metagopay] order " + orderid + " pay callback notify fail,err:" + err.Error()
		PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)

		glog.Errorf(strmsg)

		fmt.Fprint(ctx, "SC000000")
		return
	}
	strmsg := "[metagopay] order " + orderid + " pay callback notify success"
	PayPointRecord(businessid, "metagopay回调", strmsg, orderid, 1)
	fmt.Fprint(ctx, "SC000000")

	if status == 2 {
		// 订单记录
		OrederRecord(info.OrderID, info.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}
}

// metagopay提现回调
func metagopayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]interface{}) {
	status := 0
	orderid := paramMap["custOrderNo"].(string)
	order_status := paramMap["ordStatus"].(string)
	// order_time := paramMap["ordTime"].(int64)
	order_amount := fmt.Sprint(paramMap["payAmt"])
	// msg := paramMap["msg"]
	jsonStr, err := json.Marshal(paramMap)
	if err != nil {
		glog.Errorf("[metagopay] Convert to json exception,msg:%s", err.Error())
	}
	if order_status != "07" {
		var errMsg string
		if e, ok := paramMap["errMsg"]; ok {
			errMsg = e.(string)
		}
		glog.Errorf("[metagopay] order %s callback fail", orderid)
		if strings.Contains(errMsg, "Insufficient") {
			status = data.WithdrawFailTransfer
		} else if strings.Contains(errMsg, "IFSC") {
			status = data.WithdrawFailRefund
		} else {
			status = data.WithdrawFail
		}
	} else {
		status = data.WithdrawSuccess
	}

	str := string(jsonStr)
	info, err := service.GetWithDrawOrder(orderid)
	if err != nil {
		strmsg := "[metagopay] order " + orderid + " query exception,fail msg:" + err.Error()
		glog.Errorf(strmsg)
	}

	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, service.METAGOPAY, status) {
		glog.Warning("[metagopay] order %s repeat withdraw", orderid)
		fmt.Fprint(ctx, "SC000000")
		return
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == data.WithdrawSuccess {
		glog.Errorf("[metagopay] order %s is not operable", orderid)
		fmt.Fprint(ctx, "SC000000")
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

	amount, _ := strconv.Atoi(order_amount)
	order := new(data.PayNotify)
	order.OrderID = orderid
	order.MerOrderID = info.MerOrderID
	order.Status = status
	order.Amount = amount
	order.Timestamp = info.PayTime
	_, err2 := service.PayNotify(order, 2)
	if err2 != nil {
		strmsg := "[metagopay] order " + orderid + " withdraw callback notify fail,err:" + err2.Error()
		glog.Errorf(strmsg)
		fmt.Fprint(ctx, "SC000000")

		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
		return
	}
	fmt.Fprint(ctx, "SC000000")

	if status != data.WithdrawSuccess {
		tasks.CreateWithdrawSchedule(info, 0) // 轮询核单
	} else {
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		// 成功需要记录日志
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}

}
