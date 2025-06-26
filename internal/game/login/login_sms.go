package login

import (
	"github.com/valyala/fasthttp"
)

// kmi验证码回调
func kmiSms(ctx *fasthttp.RequestCtx) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		glog.Error(err)
	// 		fmt.Fprintf(ctx, "%s", "failure")
	// 	}
	// }()
	// ctx.Response.Header.Set("Content-type", "application/json")
	// switch string(ctx.Method()) {
	// case "GET":
	// default:
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }
	// clientIP := getIP(ctx)
	// glog.Infof("kmiSms clientIP: %s", clientIP)
	// //解析
	// body := ctx.QueryArgs().QueryString()
	// // rsp := make(map[string]string)
	// glog.Debug("body:", string(body))
	// param, err := smskmi.ParseCallback(body)
	// if err != nil {
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }
	// // 记录验证码
	// log := &data.LogSmsRecord{
	// 	Id:       bson.NewObjectId().String(),
	// 	SmsId:    param.SmsId,
	// 	SendTime: utils.Int64(param.SendTime),
	// 	Reslut:   param.SendResult == "SUCCESS",
	// 	SmsCount: 1,
	// 	Phone:    param.To,
	// 	State:    param.Dr,
	// 	Channel:  data.KMISMS,
	// }

	// l, err1 := json.Marshal(log)
	// if err1 != nil {
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }

	// nodePid.Tell(&pb.LogSmsRecord{Body: l})

	// // 判断是否发送成功
	// if param.SendResult == "SUCCESS" {
	// 	return
	// }
	// // 换渠道发
	// glog.Error("kmisms send fail, status：", param.Dr)
	// // 需要重发
	// msg := &pb.RetrySendSms{Phone: param.To, Stype: data.KMISMS}
	// callNode(msg)
}

// motion验证码回调
func mtSms(ctx *fasthttp.RequestCtx) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		glog.Error(err)
	// 		fmt.Fprintf(ctx, "%s", "failure")
	// 	}
	// }()
	// ctx.Response.Header.Set("Content-type", "application/json")
	// switch string(ctx.Method()) {
	// case "POST":
	// default:
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }
	// clientIP := getIP(ctx)
	// glog.Infof("mtSms clientIP: %s", clientIP)
	// //解析
	// body := ctx.PostBody()
	// rsp := make(map[string]string)
	// // glog.Infof("paynotify:%s", string(body))
	// p, err := smsmt.ParseCallback(body)
	// if err != nil {
	// 	fmt.Fprintf(ctx, "%s", "failure")
	// 	return
	// }

	// defer func() {
	// 	rsp["code"] = "00000"
	// 	r, _ := json.Marshal(rsp)
	// 	fmt.Fprintf(ctx, "%s", string(r))
	// }()

	// if p.Status != "0" {
	// 	glog.Error("[mtsms] send fail, status：", p.Status)
	// 	// 需要重发
	// 	msg := &pb.RetrySendSms{Phone: p.Phone, Stype: data.MTSMS}
	// 	_, err := callNode(msg)
	// 	if err != nil {
	// 		glog.Warningf("[mtsms] phone %s retry send fail", p.Phone)
	// 		return
	// 	}
	// }

	// // 记录验证码
	// log := &data.LogSmsRecord{
	// 	Id:       bson.NewObjectId().String(),
	// 	SmsId:    p.Uid,
	// 	Reslut:   p.Status == "0",
	// 	SmsCount: 1,
	// 	Phone:    p.Phone,
	// 	State:    p.Status,
	// 	Channel:  data.MTSMS,
	// }
	// sendTime, _ := utils.Str2Unix(p.ReportTime)
	// log.SendTime = sendTime

	// l, err1 := json.Marshal(log)
	// if err1 != nil {
	// 	glog.Warningf("[mtsms] phone %s log record fail", p.Phone)
	// 	return
	// }

	// nodePid.Tell(&pb.LogSmsRecord{Body: l})
}

// xxy验证码回调
// func xxySms(ctx *fasthttp.RequestCtx) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			glog.Error(err)
// 			fmt.Fprintf(ctx, "%s", "failure")
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
// 	glog.Infof("xxysms clientIP: %s", clientIP)
// 	//解析
// 	body := ctx.PostBody()
// 	rsp := make(map[string]string)
// 	// glog.Infof("paynotify:%s", string(body))
// 	param, err := smsxxy.ParseCallback(body)
// 	if err != nil {
// 		fmt.Fprintf(ctx, "%s", "failure")
// 		return
// 	}
// 	for _, p := range param {
// 		if p.Status != "0" {
// 			glog.Error("[xxysms] send fail, status：", p.Status)
// 			// 需要重发
// 			msg := &pb.RetrySendSms{Phone: p.Phone, Stype: data.MTSMS}
// 			_, err := callNode(msg)
// 			if err != nil {
// 				glog.Warningf("[xxysms] phone %s retry send fail", p.Phone)
// 				continue
// 			}
// 		}

// 		// 记录验证码
// 		log := &data.LogSmsRecord{
// 			Id:       bson.NewObjectId().String(),
// 			SmsId:    p.Uid,
// 			Reslut:   p.Status == "0",
// 			SmsCount: 1,
// 			Phone:    p.Phone,
// 			State:    p.Status,
// 			Channel:  data.XXYSMS,
// 		}
// 		sendTime, _ := utils.Str2Unix(p.ReportTime)
// 		log.SendTime = sendTime

// 		l, err1 := json.Marshal(log)
// 		if err1 != nil {
// 			glog.Warningf("[xxysms] phone %s log record fail", p.Phone)
// 			continue
// 		}

// 		nodePid.Tell(&pb.LogSmsRecord{Body: l})
// 	}

// 	rsp["code"] = "00000"
// 	r, _ := json.Marshal(rsp)
// 	fmt.Fprintf(ctx, "%s", string(r))
// }
