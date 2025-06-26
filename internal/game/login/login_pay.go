package login

import (
	"fmt"
	"strings"

	"goserver/gen/pb"

	"goserver/pkg/data"
	"goserver/pkg/glog"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

// 微信支付,接收交易结果通知
// func wxpayHandler(ctx *fasthttp.RequestCtx) {
// 	ctx.Response.Header.Set("Content-type", "text/plain;charset=UTF-8")
// 	switch string(ctx.Method()) {
// 	case "POST":
// 	default:
// 		glog.Error("wxpay method err")
// 		//TODO failed
// 		fmt.Fprintf(ctx, wxpay.TradeRespXml())
// 		return
// 	}
// 	result := ctx.PostBody()
// 	glog.Debugf("result %s", string(result))
// 	//解析
// 	tradeResult, err := wxpay.ParseTradeResult(result)
// 	glog.Debugf("tradeResult %#v, err %v", tradeResult, err)
// 	if err != nil {
// 		glog.Errorf("trade result err: %v", err)
// 		//TODO failed
// 		fmt.Fprintf(ctx, wxpay.TradeRespXml())
// 		return
// 	}
// 	//发货
// 	msg := new(pb.WxpayCallback)
// 	msg.Result = string(result)
// 	nodePid.Tell(msg)
// 	fmt.Fprintf(ctx, wxpay.TradeRespXml())
// }

// JTpay 竣付通支付
// var JTpay *jtpay.JTpayTrans //支付

// JtPayInit 初始化竣付通支付
// func JtPayInit() {
// 	cfg := &jtpay.JTpayConfig{
// 		Appid:         cfg.Section("jtpay").Key("yingyongnum").Value(),
// 		CompKey:       cfg.Section("jtpay").Key("compkey").Value(),
// 		PlaceOrderUrl: cfg.Section("jtpay").Key("placeorderurl").Value(),
// 	}
// 	JTpay, err = jtpay.NewJTpayTrans(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// 接收交易结果展示（同步）
// func jtpayReturn(ctx *fasthttp.RequestCtx) {
// 	ctx.Response.Header.Set("Content-type", "text/html;charset=UTF-8")
// 	switch string(ctx.Method()) {
// 	case "GET":
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	//解析
// 	param := ctx.QueryArgs().QueryString()
// 	glog.Debugf("param %s", param)
// 	tradeResult, err := jtpayNotifyVerify(param)
// 	glog.Debugf("tradeResult %#v, err %v", tradeResult, err)
// 	if tradeResult == nil || err != nil {
// 		glog.Errorf("jtpay return err: %v", err)
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	glog.Infof("jtpay return success: %#v", tradeResult)
// 	//结果页面展示
// 	result := returnHtml(tradeResult)
// 	fmt.Fprintf(ctx, "%s", result)
// 	//TODO use template
// 	//jtpayreturnPageHandler(tradeResult, ctx)
// }

// 接收交易结果通知（异步）
// func jtpayNotify(ctx *fasthttp.RequestCtx) {
// 	ctx.Response.Header.Set("Content-type", "text/plain;charset=UTF-8")
// 	switch string(ctx.Method()) {
// 	case "POST":
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	clientIP := getIP(ctx)
// 	glog.Infof("clientIP: %s", clientIP)
// 	//解析
// 	body := ctx.PostBody()
// 	glog.Debugf("body %s", string(body))
// 	tradeResult, err := jtpayNotifyVerify(body)
// 	glog.Infof("tradeResult %#v, err %v", tradeResult, err)
// 	if err != nil {
// 		glog.Errorf("jtpay notify err: %v", err)
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	//发货
// 	result, err := jsoniter.Marshal(tradeResult)
// 	if err != nil {
// 		glog.Errorf("jtpay notify err: %v", err)
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	msg := new(pb.JtpayCallback)
// 	msg.Result = result
// 	res2, err2 := callNode(msg)
// 	if err2 != nil {
// 		glog.Errorf("jtpay notify err: %v", err2)
// 	}
// 	if response2, ok := res2.(*pb.JtpayCalledback); ok {
// 		if response2.Result {
// 			glog.Infof("jtpay notify success: %#v", tradeResult)
// 		}
// 	}
// 	fmt.Fprintf(ctx, "%s", "success")
// }

// 订单验证
// func jtpayNotifyVerify(body []byte) (*jtpay.NotifyResult, error) {
// 	tradeResult, err := jtpay.ParseNotifyResult(body)
// 	if tradeResult == nil || err != nil {
// 		glog.Errorf("trade result err: %v", err)
// 		return nil, err
// 	}
// 	//支付失败
// 	if tradeResult.P4_zfstate != "1" {
// 		glog.Errorf("trade result failed: %#v", tradeResult)
// 		return nil, fmt.Errorf("trade failed %#v", tradeResult)
// 	}
// 	//参数验证
// 	if !JTpay.NotifyVerify(tradeResult) {
// 		glog.Errorf("trade result verify failed: %#v", tradeResult)
// 		return nil, fmt.Errorf("trade verify failed %#v", tradeResult)
// 	}
// 	return tradeResult, nil
// }

//请求参数
/*
p7_productcode  //商品名称,固定值WX,ZFB
p14_customname  //userid
p17_product //商品id
p19_productnum  //商品数量
p25_terminal // 终端设备类型, 1 代表 pc 2 代表 ios 3 代表 android
time //当前时间截
sign //签名
"p7_productcode=WX&p14_customname=103133&p17_product=7&p19_productnum=1&p25_terminal=2&time=11&sign=xxx"
sign=Md5(WX+1+1+2+11+key),Md5(商品名称+userid+商品id+商品数量+终端设备类型+time+key)
*/
// 接收交易下单请求
// func jtpayOrder(ctx *fasthttp.RequestCtx) {
// 	ctx.Response.Header.Set("Content-type", "text/html;charset=UTF-8")
// 	var body []byte
// 	switch string(ctx.Method()) {
// 	case "POST":
// 		body = ctx.PostBody()
// 	case "GET":
// 		body = ctx.QueryArgs().QueryString()
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	clientIP := getIP(ctx)
// 	glog.Infof("clientIP: %s", clientIP)
// 	switch clientIP {
// 	case "127.0.0.1":
// 	default:
// 		//fmt.Fprintf(ctx, "%s", "failed")
// 		//return
// 	}
// 	//解析
// 	//body := ctx.PostBody()
// 	glog.Debugf("body %s", string(body))
// 	order, err := ParseOrder(body)
// 	if order == nil || err != nil {
// 		glog.Errorf("jtpay order err %#v", err)
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	//下单
// 	err2 := jtpayOrderHandler(order, clientIP)
// 	if err2 != nil {
// 		glog.Errorf("jtpay order failed: %#v, err: %v", order, err2)
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	glog.Infof("jtpay order success: %#v", order)
// 	result := ioswap(order)
// 	glog.Debugf("result %s", result)
// 	fmt.Fprintf(ctx, "%s", result)
// 	//TODO use template
// 	//jtpayOrderPageHandler(order, ctx)
// }

// ParseOrder convert the string to struct
// func ParseOrder(resp []byte) (*jtpay.JTpayOrder, error) {
// 	order := new(jtpay.JTpayOrder)
// 	s := strings.Split(string(resp), "&")
// 	var sign, signtime string
// 	for _, v := range s {
// 		args := strings.Split(v, "=")
// 		if len(args) != 2 {
// 			return nil, errors.New("args error")
// 		}
// 		switch args[0] {
// 		case "p7_productcode":
// 			order.P7_productcode = args[1]
// 		case "p14_customname":
// 			order.P14_customname = args[1]
// 		case "p17_product":
// 			order.P17_product = args[1]
// 		case "p19_productnum":
// 			order.P19_productnum = args[1]
// 		case "p25_terminal":
// 			order.P25_terminal = args[1]
// 		case "time":
// 			signtime = args[1]
// 		case "sign":
// 			sign = args[1]
// 		default:
// 			return nil, errors.New("args error")
// 		}
// 	}
// 	//签名验证
// 	key := cfg.Section("jtpay").Key("orderkey").Value()
// 	mysign := utils.Md5(order.P7_productcode + order.P14_customname + order.P17_product +
// 		order.P19_productnum + order.P25_terminal + signtime + key)
// 	if mysign != sign {
// 		return nil, errors.New("sign failed")
// 	}
// 	return order, nil
// }

// 下单处理
// func jtpayOrderHandler(order *jtpay.JTpayOrder, ip string) error {
// //查找商品
// shop := config.GetShop(order.P17_product)
// if shop.Id == "" || shop.Payway != 1 {
// 	glog.Errorf("id err: %s", order.P17_product)
// 	return fmt.Errorf("id %s not exist", order.P17_product)
// }
// var price uint32 = shop.Price                 //RMB数量
// var diamond uint32 = shop.Number              //虚拟货币数量
// var itemid string = utils.String(shop.Propid) //货币类型
// //订单数据
// order.P3_money = utils.String(price) //RMB
// //order.P3_money = "1" //TODO test
// order.P16_customip = ip
// //order.P26_ext1 = "1.1"
// //初始化订单
// JTpay.InitOrder(order)
// glog.Infof("order %#v", order)
// //下单请求,换成页面提交
// //result, err := JTpay.Submit(order)
// //glog.Infof("result %s, err %v", result, err)
// //if err != nil {
// //	glog.Errorf("submit err %v", err)
// //	return err
// //}
// //transid,下单记录
// msg := &pb.TradeOrder{
// 	Orderid:  order.P2_ordernumber, //订单id
// 	Userid:   order.P14_customname, //玩家id
// 	Amount:   order.P19_productnum, //购买商品数量
// 	Itemid:   itemid,               //货币类型
// 	Diamond:  diamond,              //货币数量
// 	Money:    uint32(price * 100),  //转换为分
// 	Result:   data.Tradeing,        //下单状态
// 	Clientip: ip,
// }
// //请求响应
// res2, err2 := callNode(msg)
// if err2 != nil {
// 	return err2
// }
// if response2, ok := res2.(*pb.TradedOrder); ok {
// 	if response2.Result {
// 		return nil
// 	}
// }
// return fmt.Errorf("order failed %#v", msg)
// return nil
// }

// 充值埋点
func RechargeReport(code pb.RechargeReportReq_Code, content, uid string) {
	log := &pb.LogRechargeReport{
		Uid:     uid,
		Code:    int32(code),
		Content: content,
	}
	nodePid.Tell(log)
}

// 支付回调(1：代收  2:代付)
func payNotify(ctx *fasthttp.RequestCtx, rtype int) {
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
	glog.Infof("paynotify:%s", string(body))
	res := new(data.PayNotify)
	err := jsoniter.Unmarshal(body, &res)
	if err != nil {
		glog.Errorf("json unmarsals fail")
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	switch rtype {
	case 1:
		payIncome(ctx, res)
	case 2:
		payWithdraw(ctx, res)
	}
	// body := ctx.QueryArgs().QueryString()
	// var paramMap map[string]string
	// var err1 error
	// if rtype == 1 {
	// 	paramMap, err1 = xfpay.Config.ParsePayResult(body)
	// } else {
	// 	paramMap, err1 = xfpay.Config.ParseWithdrawResult(body)
	// }
	// if err1 != nil {
	// 	fmt.Fprintf(ctx, "%s", err1)
	// 	return
	// }
	// glog.Debugf("body %v", paramMap)
	// outSign := paramMap["sign"]
	// // 验证签名
	// delete(paramMap, "sign")
	// var key string = ""
	// switch rtype {
	// case 1:
	// 	key = xfpay.Config.Md5Key
	// case 2:
	// 	key = xfpay.Config.WithDrawMDd5Key
	// }
	// sign := xfpay.PaySign(paramMap, key)
	// // upper := pay.ToUpper(sign)
	// if outSign != sign {
	// 	// 签名验证失败
	// 	glog.Errorf("[xfpay]order %s verify sign fail", paramMap["merchantOrderNo"])
	// 	fmt.Fprint(ctx, "SUCCESS")
	// 	return
	// }
	// switch rtype {
	// case 1:
	// 	xfpayIncome(ctx, paramMap)
	// case 2:
	// 	xfpayWithdraw(ctx, paramMap)
	// }
}

// 充值回调
func payIncome(ctx *fasthttp.RequestCtx, notify *data.PayNotify) {
	msg := new(pb.PayOrderUpdate)
	msg.OutId = notify.OrderID
	msg.OrderId = notify.MerOrderID
	msg.State = data.TradeSuccess
	msg.Amount = int64(notify.Amount)
	msg.Otype = 1
	if notify.Status != 2 {
		msg.State = data.TradeFail
	}
	msg.OutStatus = fmt.Sprintf("%d", notify.Status)
	msg.Repair = notify.Repair

	RechargeReport(pb.PayCallback, fmt.Sprintf("order:%s, amount:%d", notify.OrderID, notify.Amount), notify.BusinessID)

	// 通知发货
	_, err := callNode(msg)
	if err != nil {
		glog.Errorf("[pay]order %s pay callback notify fail", notify.MerOrderID)
		fmt.Fprint(ctx, 0)
		return
	}
	fmt.Fprint(ctx, 0)
}

func payWithdraw(ctx *fasthttp.RequestCtx, notify *data.PayNotify) {
	msg := new(pb.PayOrderUpdate)
	msg.OutId = notify.OrderID
	msg.OrderId = notify.MerOrderID
	msg.State = int32(notify.Status)
	msg.Otype = 2
	msg.Reason = int32(notify.Reason)
	// if notify.Status != data.WithdrawSuccess {
	// 	glog.Errorf("[withdraw]order %s callback fail", notify.MerOrderID)
	// 	// msg.State = data.WithdrawFail
	// }
	msg.OutStatus = fmt.Sprintf("%d", notify.Status)
	msg.ChannelId = notify.ChannelId
	// nodePid.Tell(msg)
	// 通知发货
	_, err := callNode(msg)
	if err != nil {
		glog.Errorf("[withdraw]order %s pay callback notify fail", notify.OrderID)
		fmt.Fprint(ctx, 1)
		return
	}
	fmt.Fprint(ctx, 0)
}

func repeatWithdraw(ctx *fasthttp.RequestCtx) {
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
	glog.Infof("RepeatWithdraw:%s", string(body))
	res := new(data.RepeatWithdrawReq)
	err := jsoniter.Unmarshal(body, &res)
	if err != nil {
		glog.Errorf("json unmarsals fail")
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	msg := &pb.RepeatWithdraw{
		OrderId:   res.OrderID,
		ChannelId: res.ChannelID,
		Status:    int32(res.Status),
	}
	_, err1 := callNode(msg)
	if err1 != nil {
		glog.Errorf("[withdraw]order %s repeatwithdraw notify fail", msg.OrderId)
		fmt.Fprint(ctx, 1)
		return
	}
	fmt.Fprint(ctx, 0)
}

// ml支付回调(1：代收  2:代付)
// func mlpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			glog.Error(err)
// 		}
// 	}()
// 	ctx.Response.Header.Set("Content-type", "text/plain;charset=UTF-8")
// 	switch string(ctx.Method()) {
// 	case "GET":
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	clientIP := getIP(ctx)
// 	glog.Infof("clientIP: %s", clientIP)
// 	//解析
// 	body := ctx.QueryArgs().QueryString()
// 	glog.Debugf("body %s", string(body))
// 	paramMap := parseParam(body)
// 	outSign := paramMap["sign"]
// 	// 验证签名
// 	delete(paramMap, "sign")
// 	// msg := new(pb.PayOrderUpdate)
// 	// msg.OutId = paramMap["orderNo"]
// 	// msg.OrderId = paramMap["partnerOrderNo"]
// 	// msg.State = data.TradeSuccess
// 	var key string = ""
// 	switch rtype {
// 	case 1:
// 		key = mlpay.Config.Md5Key
// 	case 2:
// 		key = mlpay.Config.WithDrawMDd5Key
// 	}
// 	sign := mlpay.PaySign(paramMap, key)
// 	upper := pay.ToUpper(sign)
// 	if outSign != upper {
// 		// 签名验证失败
// 		glog.Errorf("[mlpay]order %s withdraw callback verify sign fail", paramMap["orderNo"])
// 		fmt.Fprint(ctx, 1)
// 		return
// 	}
// 	switch rtype {
// 	case 1:
// 		mlpayIncome(ctx, paramMap)
// 	case 2:
// 		mlpayWithdraw(ctx, paramMap)
// 	}
// }

// ml充值回调
// func mlpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	msg.OutId = paramMap["orderNo"]
// 	msg.OrderId = paramMap["partnerOrderNo"]
// 	msg.State = data.TradeSuccess
// 	msg.Otype = 1
// 	if paramMap["status"] == "0" {
// 		msg.State = data.TradeFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	// 通知发货
// 	_, err := callNode(msg)
// 	if err != nil {
// 		glog.Errorf("[mlpay]order %s pay callback notify fail", paramMap["orderNo"])
// 		fmt.Fprint(ctx, 0)
// 		return
// 	}
// 	fmt.Fprint(ctx, 0)
// }

// ml提现回调
// func mlpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	msg.OutId = paramMap["withdrawNo"]
// 	msg.OrderId = paramMap["partnerWithdrawNo"]
// 	msg.State = data.WithdrawSuccess
// 	msg.Otype = 2
// 	if paramMap["status"] == "0" {
// 		glog.Errorf("[mlpay]order %s callback fail msg:%s", paramMap["partnerWithdrawNo"], paramMap["errorMsg"])
// 		msg.State = data.WithdrawFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	nodePid.Tell(msg)
// 	// 通知发货
// 	// _, err := callNode(msg)
// 	// if err != nil {
// 	// 	glog.Errorf("[mlpay]order %s pay callback notify fail", paramMap["orderNo"])
// 	// 	fmt.Fprint(ctx, 1)
// 	// 	return
// 	// }
// 	fmt.Fprint(ctx, 0)
// }

// xf支付回调(1：代收  2:代付)
// func xfpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			glog.Error(err)
// 		}
// 	}()
// 	ctx.Response.Header.Set("Content-type", "application/json")
// 	switch string(ctx.Method()) {
// 	case "POST":
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	clientIP := getIP(ctx)
// 	glog.Infof("clientIP: %s", clientIP)
// 	//解析
// 	body := ctx.PostBody()
// 	// body := ctx.QueryArgs().QueryString()
// 	var paramMap map[string]string
// 	var err1 error
// 	if rtype == 1 {
// 		paramMap, err1 = xfpay.Config.ParsePayResult(body)
// 	} else {
// 		paramMap, err1 = xfpay.Config.ParseWithdrawResult(body)
// 	}
// 	if err1 != nil {
// 		fmt.Fprintf(ctx, "%s", err1)
// 		return
// 	}
// 	glog.Debugf("body %v", paramMap)
// 	outSign := paramMap["sign"]
// 	// 验证签名
// 	delete(paramMap, "sign")
// 	var key string = ""
// 	switch rtype {
// 	case 1:
// 		key = xfpay.Config.Md5Key
// 	case 2:
// 		key = xfpay.Config.WithDrawMDd5Key
// 	}
// 	sign := xfpay.PaySign(paramMap, key)
// 	// upper := pay.ToUpper(sign)
// 	if outSign != sign {
// 		// 签名验证失败
// 		glog.Errorf("[xfpay]order %s verify sign fail", paramMap["merchantOrderNo"])
// 		fmt.Fprint(ctx, "SUCCESS")
// 		return
// 	}
// 	switch rtype {
// 	case 1:
// 		xfpayIncome(ctx, paramMap)
// 	case 2:
// 		xfpayWithdraw(ctx, paramMap)
// 	}
// }

// xf充值回调
// func xfpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	msg.OrderId = paramMap["merchantOrderNo"]
// 	msg.State = data.TradeSuccess
// 	msg.Otype = 1
// 	if paramMap["status"] != "2" {
// 		glog.Errorf("[xfpay]order %s callback fail msg:%s", paramMap["merchantOrderNo"], paramMap["msg"])
// 		msg.State = data.TradeFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	// 通知发货
// 	_, err := callNode(msg)
// 	if err != nil {
// 		glog.Errorf("[xfpay]order %s pay callback notify fail,err:%s", paramMap["merchantOrderNo"], err)
// 		fmt.Fprint(ctx, "SUCCESS")
// 		return
// 	}
// 	fmt.Fprint(ctx, "SUCCESS")
// }

// xf提现回调
// func xfpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	// msg.OutId = paramMap["withdrawNo"]
// 	msg.OrderId = paramMap["merchantOrderNo"]
// 	msg.State = data.WithdrawSuccess
// 	msg.Otype = 2
// 	if paramMap["status"] != "2" {
// 		msg.State = data.WithdrawFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	glog.Infof("[xpay] withdraw callback, param:%v", paramMap)
// 	nodePid.Tell(msg)
// 	// 通知发货
// 	// _, err := callNode(msg)
// 	// if err != nil {
// 	// 	glog.Errorf("[mlpay]order %s pay callback notify fail", paramMap["orderNo"])
// 	// 	fmt.Fprint(ctx, 1)
// 	// 	return
// 	// }
// 	fmt.Fprint(ctx, "SUCCESS")
// }

// king支付回调(1：代收  2:代付)
// func kingpayNotify(ctx *fasthttp.RequestCtx, rtype int) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			glog.Error(err)
// 		}
// 	}()
// 	ctx.Response.Header.Set("Content-type", "application/json")
// 	switch string(ctx.Method()) {
// 	case "POST":
// 	default:
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	clientIP := getIP(ctx)
// 	glog.Infof("clientIP: %s", clientIP)
// 	//解析
// 	body := ctx.PostBody()
// 	// body := ctx.QueryArgs().QueryString()
// 	var paramMap map[string]string
// 	var err1 error
// 	if rtype == 1 {
// 		paramMap, err1 = kingpay.Config.ParsePayResult(body)
// 	} else {
// 		paramMap, err1 = kingpay.Config.ParseWithdrawResult(body)
// 	}
// 	if err1 != nil {
// 		fmt.Fprintf(ctx, "%s", err1)
// 		return
// 	}
// 	glog.Debugf("body %v", paramMap)
// 	outSign := paramMap["sign"]
// 	// 验证签名
// 	delete(paramMap, "sign")
// 	if !kingpay.Config.VerifySing(outSign, kingpay.SignContent(paramMap), crypto.SHA256) {
// 		// 签名验证失败
// 		glog.Errorf("[kingpay]order %s verify sign fail", paramMap["outOrderId"])
// 		res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "1", Message: "verify sign fail"})
// 		fmt.Fprint(ctx, res)
// 		return
// 	}
// 	switch rtype {
// 	case 1:
// 		kingpayIncome(ctx, paramMap)
// 	case 2:
// 		kingpayWithdraw(ctx, paramMap)
// 	}
// }

// king充值回调
// func kingpayIncome(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	msg.OrderId = paramMap["outOrderId"]
// 	msg.State = data.TradeSuccess
// 	msg.Otype = 1
// 	if paramMap["status"] != "4" {
// 		msg.State = data.TradeFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	// 通知发货
// 	res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "0", Message: "SUCCESS"})
// 	_, err := callNode(msg)
// 	if err != nil {
// 		glog.Errorf("[kingpay]order %s pay callback notify fail", paramMap["outOrderId"])
// 		fmt.Fprint(ctx, res)
// 		return
// 	}
// 	fmt.Fprint(ctx, res)
// }

// king提现回调
// func kingpayWithdraw(ctx *fasthttp.RequestCtx, paramMap map[string]string) {
// 	msg := new(pb.PayOrderUpdate)
// 	msg.OutId = paramMap["orderId"]
// 	msg.OrderId = paramMap["outOrderId"]
// 	msg.State = data.WithdrawSuccess
// 	msg.Otype = 2
// 	if paramMap["status"] != "5" {
// 		msg.State = data.WithdrawFail
// 	}
// 	msg.OutStatus = paramMap["status"]
// 	nodePid.Tell(msg)
// 	// 通知发货
// 	// _, err := callNode(msg)
// 	// if err != nil {
// 	// 	glog.Errorf("[mlpay]order %s pay callback notify fail", paramMap["orderNo"])
// 	// 	fmt.Fprint(ctx, 1)
// 	// 	return
// 	// }
// 	res, _ := jsoniter.Marshal(kingpay.KingPayResponse{Code: "0", Message: "SUCCESS"})
// 	fmt.Fprint(ctx, res)
// }

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

// func ioswap(order *jtpay.JTpayOrder) (str string) {
// 	str = `
// <html>
// <head>
// 	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
// 	<title>竣付通</title>
// </head>
// <!--支付宝IOSwap支付请求提交页-->
// <body onLoad="document.yeepay.submit();">
// 	<form name='yeepay' action='https://order.z.jtpay.com/jh-web-order/order/receiveOrder' method='post'  >
// 	`
// 	str += fmt.Sprintf("<input type='hidden' name='p1_yingyongnum'				value='%s'>", order.P1_yingyongnum)
// 	str += fmt.Sprintf("<input type='hidden' name='p2_ordernumber'				value='%s'>      ", order.P2_ordernumber)
// 	str += fmt.Sprintf("<input type='hidden' name='p3_money'				value='%s'>      ", order.P3_money)
// 	str += fmt.Sprintf("<input type='hidden' name='p6_ordertime'			value='%s'>  ", order.P6_ordertime)
// 	str += fmt.Sprintf("<input type='hidden' name='p7_productcode'					value='%s'>       ", order.P7_productcode)
// 	str += fmt.Sprintf("<input type='hidden' name='p8_sign'					value='%s'>       ", order.P8_sign)
// 	str += fmt.Sprintf("<input type='hidden' name='p9_signtype'			value='%s'>  ", order.P9_signtype)
// 	str += fmt.Sprintf("<input type='hidden' name='p25_terminal'			value='%s'>  ", order.P25_terminal)
// 	str += `
// 	</form>
// </body>
// </html>
// `
// 	return
// }

// func returnHtml(order *jtpay.NotifyResult) (str string) {
// 	str = `
// <!DOCTYPE html>
// <html>
// <head>`
// 	switch order.P6_productcode {
// 	case "WX":
// 		str += `<title>微信支付成功</title>`
// 	case "ZFB":
// 		str += `<title>支付宝支付成功</title>`
// 	default:
// 		str += `<title>支付成功</title>`
// 	}
// 	str += `
// <base href="/">
// <meta charset="utf-8" />
// <meta name="viewport" content="initial-scale=1.0, width=device-width, user-scalable=no" />
// <link rel="stylesheet" type="text/css" href="jtpay/css/wxzf.css">
// <script src="jtpay/js/jquery.js"></script>
// </head>
// <body >
// <div class="header">
//   <div class="all_w" style="position:relative; z-index:1;">
//     <div class="ttwenz" style=" text-align:center; width:100%;">
//       <h4>交易详情</h4>`
// 	switch order.P6_productcode {
// 	case "WX":
// 		str += `<h5>微信安全支付</h5>`
// 	case "ZFB":
// 		str += `<h5>支付宝安全支付</h5>`
// 	default:
// 		str += `<h5>安全支付</h5>`
// 	}
// 	str += `
//     </div>
//     </div>
// </div>

// <div class="zfcg_box ">
// <div class="all_w">
// <img src="jtpay/images/cg_03.jpg" > 支付成功 </div>

// </div>
// <div class="cgzf_info">
// <div class="wenx_xx">
//   <div class="mz">欢乐商城</div>`
// 	str += fmt.Sprintf("<div class='wxzf_price'>￥%s</div>", order.P3_money)
// 	str += `
// </div>

// <div class="spxx_shop">
//  <div class=" mlr_pm">

//  <table width="100%" border="0" cellspacing="0" cellpadding="0">
//   <tr>
//     <td>商   品</td>
//     <td align="right">金豆</td>
//   </tr>
//    <tr>
//     <td>交易时间</td>`
// 	str += fmt.Sprintf("<td align='right'>%s</td>", utils.Time2Str(utils.LocalTime()))
// 	str += `
//   </tr>
//    <tr>
//     <td>支付方式</td>`
// 	switch order.P6_productcode {
// 	case "WX":
// 		str += `<td align="right">微信支付</td>`
// 	case "ZFB":
// 		str += `<td align="right">支付宝支付</td>`
// 	default:
// 		str += `<td align="right">支付</td>`
// 	}
// 	str += `
//   </tr>
//    <tr>
//     <td>交易单号</td>`
// 	str += fmt.Sprintf("<td align='right'>%s</td>", order.P5_orderid)
// 	str += `
//   </tr>
// </table>

// </div>

// </div>
// </div>

// <div class="wzxfcgde_tb"><img src="jtpay/images/cg_07.jpg" ></div>

// </body>
// </html>`
// 	return
// }

// func jtpayreturnPageHandler(order *jtpay.NotifyResult, ctx *fasthttp.RequestCtx) {
// 	p := &templates.JtpayReturn{
// 		P3_money:       order.P3_money,
// 		P5_orderid:     order.P5_orderid,
// 		P6_productcode: order.P6_productcode,
// 		LocalTime:      utils.Time2Str(utils.LocalTime()),
// 	}
// 	templates.WriteJtpayReturnTemplate(ctx, p)
// 	ctx.SetContentType("text/html; charset=utf-8")
// }

// func jtpayOrderPageHandler(order *jtpay.JTpayOrder, ctx *fasthttp.RequestCtx) {
// 	p := &templates.JtpayOrder{
// 		P1_yingyongnum: order.P1_yingyongnum,
// 		P2_ordernumber: order.P2_ordernumber,
// 		P3_money:       order.P3_money,
// 		P6_ordertime:   order.P6_ordertime,
// 		P7_productcode: order.P7_productcode,
// 		P8_sign:        order.P8_sign,
// 		P9_signtype:    order.P9_signtype,
// 		P25_terminal:   order.P25_terminal,
// 	}
// 	templates.WriteJtpayOrderTemplate(ctx, p)
// 	ctx.SetContentType("text/html; charset=utf-8")
// }
