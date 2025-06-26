package pay

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/lab259/cors"
	"github.com/valyala/fasthttp"
)

// Start 启动监听服务
func Start(addr string) {
	handler := cors.Default().Handler(requestHandler)

	if err = fasthttp.ListenAndServe(addr, handler); err != nil {
		glog.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func getIP(ctx *fasthttp.RequestCtx) (ip string) {
	ip = string(ctx.Request.Header.Peek("X-Forwarded-For"))
	if ip == "" {
		ip = ctx.RemoteIP().String()
	}
	return
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		glog.Error("fail request:", r)
	// 		fmt.Fprintf(ctx, "%s", "fail")
	// 	}
	// }()
	defer ctx.SetConnectionClose()
	byteurl := ctx.Path()
	strurl := strings.TrimPrefix(string(byteurl), "/pay")
	switch strurl {
	case "/api/pay/submit":
		// 统一支付接口
		payRequestHandler(ctx)
	case "/api/withdraw/submit":
		// 统一提现接口
		withdrawRequestHandler(ctx)
	case "/api/pay/submit/channelshop": // 商品支付渠道信息查询
		payChannelShopRequestHandler(ctx)
	case "/api/syncpaychannel":
		// 支付渠道同步
		SyncPayChannel(ctx)
	case "/api/querybalance":
		// 支付渠道同步
		QueryBalance(ctx)
	case "/api/webpayrequesthandler":
		// 后台下测试单
		webpayRequestHandler(ctx)
	case "/api/pay/notify/ml":
		mlpayNotify(ctx, 1)
	case "/api/withdraw/notify/ml":
		mlpayNotify(ctx, 2)
	case "/api/pay/notify/1916":
		pay1916Notify(ctx, 1)
	case "/api/withdraw/notify/1916":
		pay1916Notify(ctx, 2)
	case "/api/pay/notify/blizzardpay":
		blizzardpayNotify(ctx, 1)
	case "/api/withdraw/notify/blizzardpay":
		blizzardpayNotify(ctx, 2)
	case "/api/pay/notify/icepay":
		icepayNotify(ctx, 1)
	case "/api/withdraw/notify/icepay":
		icepayNotify(ctx, 2)
	case "/api/pay/notify/xf":
		xfpayNotify(ctx, 1)
	case "/api/withdraw/notify/xf":
		xfpayNotify(ctx, 2)
	case "/api/pay/notify/king":
		kingpayNotify(ctx, 1)
	case "/api/withdraw/notify/king":
		kingpayNotify(ctx, 2)
	case "/api/pay/notify/sailspay":
		sailspayNotify(ctx, 1)
	case "/api/withdraw/notify/sailspay":
		sailspayNotify(ctx, 2)
	case "/api/pay/notify/flypay":
		flypayNotify(ctx, 1)
	case "/api/withdraw/notify/flypay":
		flypayNotify(ctx, 2)
	case "/api/pay/notify/xdpay":
		xdpayNotify(ctx, 1)
	case "/api/withdraw/notify/xdpay":
		xdpayNotify(ctx, 2)
	case "/api/pay/notify/uwin":
		uwinpayNotify(ctx, 1)
	case "/api/withdraw/notify/uwin":
		uwinpayNotify(ctx, 2)
	case "/api/pay/notify/kkplus":
		kkpluspayNotify(ctx, 1)
	case "/api/withdraw/notify/kkplus":
		kkpluspayNotify(ctx, 2)
	case "/api/pay/notify/tansafe":
		tansafeNotify(ctx, 1)
	case "/api/withdraw/notify/tansafe":
		tansafeNotify(ctx, 2)
	case "/api/pay/notify/letspay":
		letsPayNotify(ctx, 1)
	case "/api/withdraw/notify/letspay":
		letsPayNotify(ctx, 2)
	case "/api/pay/notify/wepay":
		wepayNotify(ctx, 1)
	case "/api/withdraw/notify/wepay":
		wepayNotify(ctx, 2)
	case "/api/pay/notify/usdt":
		usdtpayNotify(ctx, 1)
	case "/api/withdraw/notify/usdt":
		usdtpayNotify(ctx, 2)
	case "/api/pay/notify/oepay":
		oepayNotify(ctx, 1)
	case "/api/withdraw/notify/oepay":
		oepayNotify(ctx, 2)
	case "/api/pay/notify/9spay":
		pay9sNotify(ctx, 1)
	case "/api/withdraw/notify/9spay":
		pay9sNotify(ctx, 2)
	case "/api/pay/notify/metagopay":
		metagopayNotify(ctx, 1)
	case "/api/withdraw/notify/metagopay":
		metagopayNotify(ctx, 2)
	case "/api/pay/notify/universalpay":
		universalpayNotify(ctx, 1)
	case "/api/withdraw/notify/universalpay":
		universalpayNotify(ctx, 2)
	default:
		mjlNotify(ctx, string(strurl))
	}
}

func mjlNotify(ctx *fasthttp.RequestCtx, strurl string) {
	switch strurl {
	case "/api/pay/notify/universe":
		uniPayNotify(ctx, 1)
	case "/api/withdraw/notify/universe":
		uniPayNotify(ctx, 2)
	case "/api/pay/notify/jypay":
		jyPayNotify(ctx, 1)
	case "/api/withdraw/notify/jypay":
		jyPayNotify(ctx, 2)
	case "/api/pay/notify/99pay":
		_99PayNotify(ctx, 1)
	case "/api/withdraw/notify/99pay":
		_99PayNotify(ctx, 2)
	case "/api/pay/notify/transafepay":
		transafePayNotify(ctx, 1)
	case "/api/withdraw/notify/transafepay":
		transafePayNotify(ctx, 2)
	case "/api/pay/notify/mmpay":
		mmPayNotify(ctx, 1)
	case "/api/withdraw/notify/mmpay":
		mmPayNotify(ctx, 2)
	case "/api/pay/notify/safepay":
		safePayNotify(ctx, 1)
	case "/api/withdraw/notify/safepay":
		safePayNotify(ctx, 2)
	case "/api/pay/notify/gopay":
		goPayNotify(ctx, 1)
	case "/api/withdraw/notify/gopay":
		goPayNotify(ctx, 2)
	case "/api/pay/notify/shpay":
		shPayNotify(ctx, 1)
	case "/api/withdraw/notify/shpay":
		shPayNotify(ctx, 2)
	case "/api/pay/notify/mjlblizzardpay":
		mjlblizzardpayNotify(ctx, 1)
	case "/api/withdraw/notify/mjlblizzardpay":
		mjlblizzardpayNotify(ctx, 2)
	default:
		payCallback(ctx, strurl)
	}
}

// 解析回调路径，返回渠道名和回调类型(1:充值 2:提现)
func ParseCallbackPath(path string) (callback service.IPayCallback, callbackType int, err error) {
	// 支付回调路径前缀
	const (
		payPrefix      = "/api/pay/notify/"
		withdrawPrefix = "/api/withdraw/notify/"
	)

	switch {
	case strings.HasPrefix(path, payPrefix):
		callbackType = 1
		callback = service.GetCallback(path[len(payPrefix):])
	case strings.HasPrefix(path, withdrawPrefix):
		callbackType = 2
		callback = service.GetCallback(path[len(withdrawPrefix):])
	default:
		return nil, 0, fmt.Errorf("invalid callback path: %s", path)
	}

	if callback == nil {
		return nil, 0, fmt.Errorf("callback not found")
	}

	return
}

func handleError(ctx *fasthttp.RequestCtx, err error) {
	if err != nil {
		glog.Error(err)
		fmt.Fprintf(ctx, "%s", "fail")
	}
}

// 支付回调
func payCallback(ctx *fasthttp.RequestCtx, strurl string) {
	defer func() {
		if err := recover(); err != nil {
			handleError(ctx, fmt.Errorf("%v", err))
		}
	}()

	callback, callbackType, err := ParseCallbackPath(strurl)
	if err != nil {
		handleError(ctx, err)
		return
	}

	if !callback.CheckMethod(ctx) {
		handleError(ctx, fmt.Errorf("invalid request method"))
		return
	}

	paramMap, err := callback.ParseParamMap(ctx)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// 验证签名
	ok, err := callback.VerifySign(paramMap)
	if err != nil {
		handleError(ctx, err)
		return
	}

	if !ok {
		handleError(ctx, fmt.Errorf("verify sign error"))
		return
	}

	if callbackType == 1 {
		err = payCallbackIncome(ctx, paramMap, callback)
		if err != nil {
			handleError(ctx, err)
		}
	} else if callbackType == 2 {
		if service.LEOPAY == callback.GetChannelId() {
			// leopay 提现回调延迟执行
			time.Sleep(5 * time.Second)
		}

		err = payCallbackWithdraw(ctx, paramMap, callback)
		if err != nil {
			handleError(ctx, err)
		}
	} else {
		handleError(ctx, fmt.Errorf("invalid callback type"))
	}
}

func payCallbackIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string, callback service.IPayCallback) error {
	paramInfo, err := callback.ParseIncomeParamInfo(paramMap)
	if err != nil {
		return err
	}

	milliseconds := utils.BsonNow().UnixNano() / 1e6

	info, err := service.GetPayOrder(paramInfo.OrderId)
	if err != nil {
		return err
	}

	// 判断订单状态是否可操作
	if info.OrderStatus == 2 || info.ChannelId != callback.GetChannelId() {
		glog.Errorf("[%s] order %s is not operable,The order status is %d", callback.GetChannelName(), paramInfo.OrderId, info.OrderStatus)
		fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())
		return nil
	}

	info.OutTradeStatus = paramInfo.Status
	info.OutTradeNo = paramInfo.TradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = paramInfo.JsonStr
	info.OrderStatus = int32(paramInfo.StatusCode)
	info.RefuseReason = paramInfo.Message

	err = service.UpdateOrder(info)
	if err != nil {
		return fmt.Errorf("[%s] order %s update order error,msg:%s", callback.GetChannelName(), paramInfo.OrderId, err.Error())
	}

	// 通知服务器
	order := new(data.PayNotify)
	order.OrderID = paramInfo.OrderId
	order.MerOrderID = info.MerOrderID
	order.Status = paramInfo.StatusCode
	order.Amount = paramInfo.Amount
	order.Timestamp = int64(milliseconds)

	_, err = service.PayNotify(order, 1)
	if err != nil {
		glog.Errorf("[%s] order %s pay notify error,msg:%s", callback.GetChannelName(), paramInfo.OrderId, err.Error())
		fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())
		return nil
	}

	fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())

	if paramInfo.StatusCode == 2 {
		OrederRecord(info.OrderID, info.ChannelId, 2)
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 1, info.ChannelId, int64(info.Amount))
	}

	return nil
}

func payCallbackWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string, callback service.IPayCallback) error {
	paramInfo, err := callback.ParseWithdrawParamInfo(paramMap)
	if err != nil {
		return err
	}

	milliseconds := utils.BsonNow().UnixNano() / 1e6

	info, err := service.GetWithDrawOrder(paramInfo.OrderId)
	if err != nil {
		return fmt.Errorf("[%s] order %s get withdraw order error,msg:%s", callback.GetChannelName(), paramInfo.OrderId, err.Error())
	}

	WithdrawOrderLock(paramInfo.OrderId)
	// 顺序不能乱
	defer WithdrawOrderUnlock(paramInfo.OrderId, true)
	defer service.UpdateWithDrawOrder(info)

	// 是否重复付款
	if IsRepeatPay(info, callback.GetChannelId(), paramInfo.StatusCode) {
		glog.Errorf("[%s] order %s is repeat withdraw", callback.GetChannelName(), paramInfo.OrderId)
		fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())
		return nil
	}

	//判断订单状态是否可操作
	if info.OrderStatus == 2 {
		glog.Errorf("[%s] order %s is not operable", callback.GetChannelName(), paramInfo.OrderId)
		fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())
		return nil
	}

	info.OutTradeStatus = paramInfo.Status
	info.OutTradeNo = paramInfo.TradeNo
	info.PayTime = int64(milliseconds)
	info.CallBackMsg = paramInfo.JsonStr
	info.OrderStatus = int32(paramInfo.StatusCode)
	info.RefuseReason = paramInfo.Message

	//通知服务器
	order := new(data.PayNotify)
	order.OrderID = paramInfo.OrderId
	order.MerOrderID = info.MerOrderID
	order.Status = paramInfo.StatusCode
	order.Amount = paramInfo.Amount
	order.ChannelId = callback.GetChannelId()
	order.Timestamp = int64(milliseconds)

	_, err = service.PayNotify(order, 2)
	if err != nil {
		glog.Errorf("[%s] order %s withdraw notify error,msg:%s", callback.GetChannelName(), paramInfo.OrderId, err.Error())
		fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())
		// tasks.CreateWithdrawSchedule(info, 0)
		return nil
	}

	fmt.Fprintf(ctx, "%s", callback.GetSuccessMsg())

	if paramInfo.StatusCode != data.WithdrawSuccess {
		tasks.CreateWithdrawSchedule(info, 0)
	} else {
		WithdrawRecord(order.OrderID, order.ChannelId, 2)
		tasks.CreatePayLogSchedule(info.OrderID, info.Userid, 2, info.ChannelId, int64(info.Amount))
	}

	return nil
}

// 代收请求
func payRequestHandler(ctx *fasthttp.RequestCtx) {
	// 查询所有可使用支付
	//list := data.GetPayChannelList()
	rsp := new(data.PayResponse)
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		rsp.Code = 1
		rsp.Msg = "failure"
		returnPayMsg(ctx, rsp)
		return
	}
	//解析
	body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	req := new(data.PayRequest)
	err := jsoniter.Unmarshal(body, &req)
	if err != nil {
		msgstr := "代收请求参数解析失败,msg:" + err.Error()
		// 埋点数据上报
		PayPointRecord(req.BusinessID, "代收下单", msgstr, string(body), 1)
		rsp.Code = 2
		rsp.Msg = msgstr
		returnPayMsg(ctx, rsp)
		return
	}
	if req.OrderID == "" || req.Amount <= 0 {
		rsp.Code = 2
		rsp.Msg = "请求参数不正确！"
		// 埋点数据上报
		PayPointRecord(req.BusinessID, "代收下单", "代付请求参数不正确！", string(body), 1)
		returnPayMsg(ctx, rsp)
		return
	} else {
		if req.Amount < 10000 {
			// 埋点数据上报
			PayPointRecord(req.BusinessID, "代收下单", "金额不能小于100", string(body), 1)
			rsp.Code = 2
			rsp.Msg = "金额不能小于100"
			returnPayMsg(ctx, rsp)
			return
		}
	}
	// 埋点数据上报
	PayPointRecord(req.BusinessID, "代收下单", "解析参数成功", string(body), 1)
	order, err1 := service.CreateOrder(*req)
	if err1 != nil {
		// 埋点数据上报
		msgstr := "创建代收订单失败,msg:" + err1.Error()
		PayPointRecord(req.BusinessID, "代收下单", msgstr, string(body), 1)
		rsp.Code = 3
		rsp.Msg = "创建代收订单失败！"
		returnPayMsg(ctx, rsp)
		return
	} else {
		// 埋点数据上报
		orderinfo, _ := jsoniter.Marshal(order)
		PayPointRecord(req.BusinessID, "代收下单", "代收订单创建", string(orderinfo), 1)
	}
	order.PlaceAnOrder = string(body) // 记录下单参数
	rsp.OrderID = order.OrderID
	rsp.ChannelID = order.ChannelId
	rsp.MerOrderID = order.MerOrderID

	// 开始下单
	exclude := make([]string, 0)
	RetryPayOrder(ctx, order, rsp, exclude, false)

	// 埋点数据上报
	res, _ := jsoniter.Marshal(rsp)
	PayPointRecord(req.BusinessID, "代收下单", "代收下单响应", string(res), 1)
	returnPayMsg(ctx, rsp)

}

// 后台测试代收请求
func webpayRequestHandler(ctx *fasthttp.RequestCtx) {
	// 查询所有可使用支付
	//list := data.GetPayChannelList()
	rsp := new(data.PayResponse)
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		rsp.Code = 1
		rsp.Msg = "failure"
		returnPayMsg(ctx, rsp)
		return
	}
	//解析
	body := ctx.PostBody()
	// body := ctx.QueryArgs().QueryString()
	req := new(data.WebPayRequest)
	err := jsoniter.Unmarshal(body, &req)
	if err != nil {
		msgstr := "代收请求参数解析失败,msg:" + err.Error()
		rsp.Code = 2
		rsp.Msg = msgstr
		returnPayMsg(ctx, rsp)
		return
	}
	if req.OrderID == "" || req.Amount <= 0 {
		rsp.Code = 2
		rsp.Msg = "请求参数不正确！"
		returnPayMsg(ctx, rsp)
		return
	}
	if req.ChannelId == 0 {
		rsp.Code = 2
		rsp.Msg = "通道参数不正确！"
		returnPayMsg(ctx, rsp)
		return
	}
	order, err1 := service.CreateWebOrder(*req)
	if err1 != nil {
		rsp.Code = 3
		rsp.Msg = "创建代收订单失败！"
		returnPayMsg(ctx, rsp)
		return
	}

	order.PlaceAnOrder = string(body) // 记录下单参数
	rsp.OrderID = order.OrderID
	rsp.ChannelID = order.ChannelId
	rsp.MerOrderID = order.MerOrderID

	// 开始下单
	channelMent := service.GetChannelMent(req.ChannelId)
	if channelMent == nil {
		rsp.Code = 3
		rsp.Msg = "通道不存在！"
		returnPayMsg(ctx, rsp)
		return
	}

	start := time.Now()
	res, err := channelMent.Submit(order) // 下单

	// 更新订单信息
	rsp.ChannelID = req.ChannelId
	order.RequestMsg = string(res)
	order.ChannelId = req.ChannelId

	defer service.UpdateOrder(order)

	if err != nil {
		rsp.Code = 3
		rsp.Msg = "提单请求失败: " + err.Error()
		returnPayMsg(ctx, rsp)
		return
	}

	elapsed := time.Since(start)
	glog.Infof("下单响应处理耗时:%s", elapsed)
	strurl, err := channelMent.SubmitResponse(res, order)
	if err != nil {
		rsp.Code = 3
		rsp.Msg = "提单失败: " + err.Error()
		returnPayMsg(ctx, rsp)
		return
	}

	if strurl != "" {
		rsp.Code = 200
		rsp.PaymentUrl = strurl
	} else {
		// 没返回收银台地址
		rsp.Code = 3
		rsp.Msg = "未返回收银台地址"
	}
	returnPayMsg(ctx, rsp)
}

// 代付请求
func withdrawRequestHandler(ctx *fasthttp.RequestCtx) {
	// 查询所有可使用支付
	//list := data.GetPayChannelList()
	rsp := new(data.WithdrawResponse)
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		rsp.Code = 1
		rsp.Msg = "failure"
		returnMsg(ctx, rsp)
		// fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	//解析
	body := ctx.PostBody()
	req := new(data.WithdrawRequest)

	err := jsoniter.Unmarshal(body, &req)
	if err != nil {
		msgstr := "代付请求参数解析失败,msg:" + err.Error()
		rsp.Code = 100
		rsp.Msg = msgstr
		returnMsg(ctx, rsp)
		// fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	if req.OrderID == "" || req.Amount <= 0 {
		rsp.Code = 100
		rsp.Msg = "请求参数不正确！"
		returnMsg(ctx, rsp)
		return
	} else {
		if req.Amount < 10000 {
			rsp.Code = 100
			rsp.Msg = "金额不能小于100"
			returnMsg(ctx, rsp)
			return
		}
	}

	var err1 error
	var order *entity.WithdrawOrder
	if req.Transfer {
		// 转单的，不生成新订单
		order, err1 = service.GetWithdrawByMerOrder(req.OrderID)
	} else {
		order, err1 = service.CreateWithDrawOrder(*req)
	}
	if err1 != nil {
		// 埋点数据上报
		msgstr := "创建代付订单失败,msg:" + err1.Error()
		rsp.Code = 100
		rsp.Msg = msgstr
		returnMsg(ctx, rsp)
		return
		// fmt.Fprintf(ctx, "%s", "failure")
		// return
	}

	// 加锁，防止回调比响应处理还快
	WithdrawOrderLock(order.OrderID)
	defer WithdrawOrderUnlock(order.OrderID, false)

	tryToWithdrawSubmit(order, rsp)

	rsp.OrderID = order.OrderID
	rsp.ChannelID = order.ChannelId
	rsp.MerOrderID = order.MerOrderID
	rsp.Status = order.OrderStatus
	returnMsg(ctx, rsp)
}

func tryToWithdrawSubmit(order *entity.WithdrawOrder, rsp *data.WithdrawResponse) {
	// channelId := service.GetChannelId(false, order.UsedChannelId, order.PayWay)
	channelId := service.GetWithdrawChannelId(order.UsedChannelId, order.PayWay, int64(order.Amount), order.Country, order.Bank)
	if channelId == 0 { // TODO 没渠道可用，退钱
		// order.OrderStatus = 3
		order.OrderStatus = data.DeliveryFailRefund
		service.UpdateWithDrawOrder(order)
		msgstr := "no can use channel"
		rsp.Code = 100
		rsp.Msg = msgstr
		return
	}

	defer service.UpdateWithDrawOrder(order)
	order.UsedChannelId = append(order.UsedChannelId, strconv.Itoa(int(channelId)))
	order.ChannelId = channelId
	withdrawMent := service.GetChannelMent(order.ChannelId)
	if withdrawMent == nil {
		glog.Errorf("order:%s no found channel:%d", order.OrderID, order.ChannelId)
		// if true { // TODO 自动转单
		// 	tryToWithdrawSubmit(order, rsp)
		// } else {
		order.OrderStatus = data.DeliveryFailTransfer
		// service.UpdateWithDrawOrder(order)
		rsp.Code = 100
		rsp.Msg = "can not use channel"
		// }
		return
	}

	res, err := withdrawMent.WithdrawSubmit(order)
	// 更新订单信息
	order.RequestMsg = string(res)
	if err != nil {
		rsp.Code = 100
		rsp.Msg = "代付订单请求下单失败,msg:" + err.Error()
		// if true { // TODO 自动转单
		// 	tryToWithdrawSubmit(order, rsp)
		// } else {
		order.OrderStatus = data.DeliveryFailTransfer
		// service.UpdateWithDrawOrder(order)
		rsp.Code = 100
		rsp.Msg = "代付订单请求下单失败"
		// }
		return
	}

	isOrder, msg := withdrawMent.WithdrawSubmitResponse(res, order)
	if isOrder {
		rsp.Code = 200
		// 订单记录
		WithdrawRecord(order.OrderID, order.ChannelId, 1)
	} else {
		// order.OrderStatus = data.OrderFail
		rsp.Code = 100
		rsp.Msg = msg
	}

	// if rsp.Code != 200 {
	// 	// 换渠道
	// 	glog.Debugf("withdraw submit fail, channel:%d, orderid:%s", order.ChannelId, order.OrderID)
	// 	tryToWithdrawSubmit(order, rsp)
	// 	return
	// }
}

func returnPayMsg(ctx *fasthttp.RequestCtx, rsp *data.PayResponse) {
	res, _ := jsoniter.Marshal(rsp)
	fmt.Fprint(ctx, string(res))
}

func returnMsg(ctx *fasthttp.RequestCtx, rsp *data.WithdrawResponse) {
	res, _ := jsoniter.Marshal(rsp)
	fmt.Fprint(ctx, string(res))
}

/*
// 埋点数据上报

*/

func PayPointRecord(businessid, name, msg, body string, types int) {
	go func() {
		pointinfo := new(entity.PayPointRecord)
		pointinfo.BusinessID = businessid
		pointinfo.Types = types
		pointinfo.Module = name
		pointinfo.Msg = msg
		pointinfo.Body = body
		err := service.AddPayPointRecord(pointinfo)
		if err != nil {
			glog.Errorf("[PayPointRecord] Failed to add new record,err:%s", err.Error())
		}
	}()
}

// 上一单下单失败，重试
func RetryPayOrder(ctx *fasthttp.RequestCtx, order *entity.PayOrder, rsp *data.PayResponse, exclude []string, retry bool) {
	var channel uint32 = 0
	// 是否是重试
	if retry {
		// channel = service.GetChannelId(true, exclude, order.PayWay)
		channel = service.GetPayChannelId(exclude, order.PayWay, order.PayOption, order.PayApp, int64(order.Amount))
	} else {
		channel = order.ChannelId
	}

	if channel == 0 {
		// 没有可用通道了，这单下单失败
		msgstr := "代收订单请求下单失败"
		// PayPointRecord(req.BusinessID, "代收下单", msgstr, string(res), 1)
		rsp.Code = 100
		rsp.Msg = msgstr
		// returnPayMsg(ctx, rsp)
		return
	}
	channelMent := service.GetChannelMent(channel)
	if channelMent == nil {
		// 下单失败通知
		publishPayOrderError(channel, order.Userid, order.MerOrderID, order.OrderID, order.Amount, "通道接口未实现", "")
		// 继续重试
		glog.Warningf("channelMent %d is nil，retry", channel)
		exclude = append(exclude, utils.String(channel))
		RetryPayOrder(ctx, order, rsp, exclude, true)
		return
	}

	start := time.Now()
	res, err := channelMent.Submit(order) // 下单

	// 更新订单信息
	rsp.ChannelID = channel
	order.RequestMsg = string(res)
	order.ChannelId = channel

	defer service.UpdateOrder(order)

	if err != nil {
		// 下单失败通知
		publishPayOrderError(channel, order.Userid, order.MerOrderID, order.OrderID, order.Amount, "提单失败:"+err.Error(), "")

		// 继续重试
		glog.Warningf("Submit err,retry,nowchannel:%d,err:%s", channel, err)
		exclude = append(exclude, utils.String(channel))
		if order.PayWay != 1 {
			RetryPayOrder(ctx, order, rsp, exclude, true)
		}
		return
	}
	elapsed := time.Since(start)
	glog.Infof("下单响应处理耗时:%s", elapsed)
	strurl, err := channelMent.SubmitResponse(res, order)
	if err != nil {
		// 下单失败通知
		publishPayOrderError(channel, order.Userid, order.MerOrderID, order.OrderID, order.Amount, "提单响应失败:"+err.Error(), order.RequestMsg)

		// 继续重试
		glog.Warningf("SubmitResponse err,retry,nowchannel:%d,err:%s", channel, err)
		exclude = append(exclude, utils.String(channel))
		if order.PayWay != 1 {
			RetryPayOrder(ctx, order, rsp, exclude, true)
		}
		return
	}
	if strurl != "" {
		rsp.Code = 200
		rsp.PaymentUrl = strurl
		// 订单记录
		OrederRecord(order.OrderID, order.ChannelId, 1)
	} else {
		// 下单失败通知
		publishPayOrderError(channel, order.Userid, order.MerOrderID, order.OrderID, order.Amount, "未返回收银台地址", order.RequestMsg)

		// 没返回收银台地址,继续重试
		glog.Warningf("收银台地址错误，重试,nowchannel:%d", channel)
		exclude = append(exclude, utils.String(channel))
		if order.PayWay != 1 {
			RetryPayOrder(ctx, order, rsp, exclude, true)
		}
		return
	}
}

// 下单失败通知MQ
func publishPayOrderError(channelId uint32, userid, merOrderId, orderId string, amount uint32, err, response string) {
	channel := service.GetPayChannel(utils.String(channelId))
	mq.NatsPublish(mq.TopicPayOrderError, &mq.PublishPayOrderError{
		ChannelId:   channelId,
		ChannelName: channel.Name,
		Userid:      userid,
		MerOrderId:  merOrderId,
		OutTradeNo:  orderId,
		Amount:      amount,
		Err:         err,
		Response:    response,
	})
}

func SyncPayChannel(ctx *fasthttp.RequestCtx) {
	rsp := &entity.WebResponse{Code: 200}
	defer webReponse(ctx, rsp)

	switch string(ctx.Method()) {
	case "POST":
	default:
		rsp.Code = -1
		rsp.ErrMsg = "must POST"
		return
	}

	//解析
	body := ctx.PostBody()

	datas := make([]entity.PayChannel, 0)

	err := jsoniter.Unmarshal(body, &datas)
	if err != nil {
		// fmt.Fprintf(ctx, "%s", "Unmarshal paychannel fail")
		rsp.Code = -1
		rsp.ErrMsg = "Unmarshal paychannel fail"
		return
	}

	service.SetPayChannel(datas)
	rsp.Code = 200

	// 通知 gate 支付渠道更新
	pub := new(pb.PublishPayChannelUpdate)
	for _, data := range datas {
		pub.Channels = append(pub.Channels, data.Id)
	}
	mq.NatsPublish(mq.TopicPaychannelUpdate, pub)
}

// 余额查询
func QueryBalance(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "must GET")
		return
	}

	response := &entity.WebResponse{Code: 200}

	defer webReponse(ctx, response)

	//解析
	body := ctx.QueryArgs().Peek("channelId")
	str := string(body)

	channel, _ := strconv.Atoi(str)

	if channel <= 0 {
		wg := new(sync.WaitGroup)
		wg.Add(len(service.ChannelRegister))
		// 查询所有渠道
		bsc := make(chan entity.WebQueryBalence, len(service.ChannelRegister))
		for k, c := range service.ChannelRegister {
			go func(k uint32, c service.IPay) {
				defer wg.Done()
				glog.Infof("%d balance query", k)
				balence := c.BalanceQuery()
				bsc <- entity.WebQueryBalence{ChannelId: k, Balence: balence}
			}(k, c)
		}
		wg.Wait()

		balences := make([]entity.WebQueryBalence, 0)
	loop:
		for {
			select {
			case bs := <-bsc:
				balences = append(balences, bs)
			default:
				close(bsc)
				break loop
			}
		}

		// balences := make([]entity.WebQueryBalence, 0)
		// for k, c := range service.ChannelRegister {
		// 	balence := c.BalanceQuery()
		// 	balences = append(balences, entity.WebQueryBalence{ChannelId: k, Balence: balence})
		// }
		bytes, err := jsoniter.Marshal(balences)
		if err != nil {
			response.Code = -1
			response.ErrMsg = "Marshal balance fail"
			return
		}
		response.Body = bytes
		return
	}

	if c, ok := service.ChannelRegister[uint32(channel)]; ok {
		balence := c.BalanceQuery()
		bytes, err := jsoniter.Marshal(entity.WebQueryBalence{ChannelId: uint32(channel), Balence: balence})
		if err != nil {
			response.Code = -1
			response.ErrMsg = "Marshal balance fail"
			return
		}
		response.Body = bytes
		return
	}
}

func UtrOrder(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "must GET")
		return
	}

	response := &entity.WebResponse{Code: 200}

	defer webReponse(ctx, response)

	//解析
	channelId := string(ctx.QueryArgs().Peek("channelId"))
	orderId := string(ctx.QueryArgs().Peek("orderId"))
	utr := string(ctx.QueryArgs().Peek("utr"))

	if channelId == "" || orderId == "" || utr == "" {
		response.Code = -1
		response.ErrMsg = "param error"
		return
	}

	channel := service.GetCallback(channelId)
	if channel == nil {
		response.Code = -1
		response.ErrMsg = "channel not found"
		return
	}

}

func webReponse(ctx *fasthttp.RequestCtx, response *entity.WebResponse) {
	rsp, _ := jsoniter.Marshal(response)
	fmt.Fprintf(ctx, "%s", string(rsp))
}

func OrederRecord(orderId string, channelId uint32, state int) {
	service.Mutex.Lock()
	defer service.Mutex.Unlock()

	history := &service.PayChannelHistory{
		ChannelId:   channelId,
		Info:        make([]*service.PayInfo, 0),
		SuccessRate: 10000,
	}

	exists := false
	for _, p := range service.PayHistory {
		if p.ChannelId == channelId {
			exists = true
			history = p
		}
	}

	add := true
	for _, i := range history.Info {
		if i.OrderID == orderId {
			add = false
			i.Success = state == 2
			break
		}
	}
	if add {
		history.Info = append(history.Info, &service.PayInfo{
			OrderID: orderId,
			Success: state == 2,
		})
		if len(history.Info) > 100 {
			history.Info = history.Info[len(history.Info)-100:]
		}
	}

	if !exists {
		service.PayHistory = append(service.PayHistory, history)
	}

	if len(history.Info) < 100 {
		history.SuccessRate = 10000
		return
	}

	// 计算成功率
	success := 0
	for _, info := range history.Info {
		if info.Success {
			success++
		}
	}

	history.SuccessRate = success * 10000 / len(history.Info)

	// 排序
	sort.Slice(service.PayHistory, func(i, j int) bool {
		return service.PayHistory[i].SuccessRate > service.PayHistory[j].SuccessRate
	})
}

func WithdrawRecord(orderId string, channelId uint32, state int) {
	service.WithdrawMutex.Lock()
	defer service.WithdrawMutex.Unlock()

	history := &service.PayChannelHistory{
		ChannelId:   channelId,
		Info:        make([]*service.PayInfo, 0),
		SuccessRate: 10000,
	}

	exists := false
	for _, p := range service.WithdrawHistory {
		if p.ChannelId == channelId {
			exists = true
			history = p
		}
	}

	add := true
	for _, i := range history.Info {
		if i.OrderID == orderId {
			add = false
			i.Success = state == 2
			break
		}
	}
	if add {
		history.Info = append(history.Info, &service.PayInfo{
			OrderID: orderId,
			Success: state == 2,
		})
		if len(history.Info) > 20 {
			history.Info = history.Info[len(history.Info)-20:]
		}
	}

	if !exists {
		service.WithdrawHistory = append(service.WithdrawHistory, history)
	}

	if len(history.Info) < 20 {
		history.SuccessRate = 10000
		return
	}

	// 计算成功率
	success := 0
	for _, info := range history.Info {
		if info.Success {
			success++
		}
	}

	history.SuccessRate = success * 10000 / len(history.Info)

	// 排序
	sort.Slice(service.WithdrawHistory, func(i, j int) bool {
		return service.WithdrawHistory[i].SuccessRate > service.WithdrawHistory[j].SuccessRate
	})
}

func returnJson(ctx *fasthttp.RequestCtx, rsp any) {
	res, _ := jsoniter.Marshal(rsp)
	fmt.Fprint(ctx, string(res))
}

// 商品支付渠道查询
func payChannelShopRequestHandler(ctx *fasthttp.RequestCtx) {
	rsp := new(data.PayChannelShopResponse)
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "POST":
	default:
		rsp.Code = 1
		rsp.Msg = "failure"
		returnJson(ctx, rsp)
		return
	}

	// 查询所有可使用支付
	service.PayMap.Range(func(_, c any) bool {
		payChannel := c.(entity.PayChannel)
		payable := payChannel.Status == 1 && payChannel.PayWeight > 0
		withdrawable := payChannel.Wstatus == 1 && payChannel.WithdrawWeight > 0
		if !payable && !withdrawable {
			return true
		}
		rsp.Channels = append(rsp.Channels, data.PayChannelShop{
			Id:              payChannel.Id,
			PayOptions:      payChannel.PayOptions,
			PayApps:         payChannel.PayApps,
			UtrRequired:     payChannel.UtrRequired,
			Payable:         payable,
			PayMin:          payChannel.PayMin,
			PayMax:          payChannel.PayMax,
			Withdrawable:    withdrawable,
			WithdrawMin:     payChannel.WithdrawMin,
			WithdrawMax:     payChannel.WithdrawMax,
			WithdrawBanks:   payChannel.WithdrawBanks,
			WithdrawWallets: payChannel.WithdrawWallets,
		})
		return true
	})

	returnJson(ctx, rsp)
}
