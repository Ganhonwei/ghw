package sms

import (
	"encoding/json"
	"fmt"
	"goserver/pkg/sms/smscl"
	"goserver/pkg/sms/smskmi"
	"goserver/pkg/sms/smsmt"
	"goserver/pkg/sms/smsxxy"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/valyala/fasthttp"
	"gopkg.in/mgo.v2/bson"
)

func writeJSON(ctx *fasthttp.RequestCtx, code int, data interface{}) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte(`{"code":500,"message":"json marshal error"}`))
		return
	}

	ctx.SetStatusCode(code)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonBytes)
}

func sendSMSCode(ctx *fasthttp.RequestCtx) {
	type SmsRequest struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	type SmsResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	defer func() {
		if err := recover(); err != nil {
			writeJSON(ctx, fasthttp.StatusInternalServerError, SmsResponse{
				Code:    500,
				Message: fmt.Sprintf("系统错误: %v", err),
			})
		}
	}()

	if !ctx.IsPost() {
		writeJSON(ctx, fasthttp.StatusMethodNotAllowed, SmsResponse{
			Code:    405,
			Message: "方法错误",
		})
		return
	}

	body := ctx.PostBody()

	var request SmsRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, SmsResponse{
			Code:    400,
			Message: fmt.Sprintf("请求参数错误: %v", err),
		})
		return
	}

	msg := &pb.SendSmsCode{
		Phone: request.Phone,
		Code:  request.Code,
	}
	nodePid.Tell(msg)

	writeJSON(ctx, fasthttp.StatusOK, SmsResponse{
		Code:    200,
		Message: "发送成功",
	})
}

// kmi验证码回调
func kmiSms(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "failure")
		}
	}()
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("kmiSms clientIP: %s", clientIP)
	//解析
	body := ctx.QueryArgs().QueryString()
	// rsp := make(map[string]string)
	glog.Debug("body:", string(body))
	param, err := smskmi.ParseCallback(body)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	// 记录验证码
	log := &data.LogSmsRecord{
		Id:       bson.NewObjectId().String(),
		SmsId:    param.SmsId,
		SendTime: utils.Int64(param.SendTime),
		Reslut:   param.SendResult == "SUCCESS",
		SmsCount: 1,
		Phone:    param.To,
		State:    param.Dr,
		Channel:  data.KMISMS,
	}
	log.Save()

	// 判断是否发送成功
	if param.SendResult == "SUCCESS" {
		return
	}
	// 换渠道发
	glog.Errorf("kmisms send fail, status：%s,phone:%s", param.Dr, param.To)
	// 需要重发
	msg := &pb.RetrySendSms{Phone: param.To[2:], Stype: data.KMISMS}
	nodePid.Tell(msg)
}

// motion验证码回调
func mtSms(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "failure")
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
	glog.Infof("mtSms clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	rsp := make(map[string]string)
	// glog.Infof("paynotify:%s", string(body))
	params, err := smsmt.ParseCallback(body)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	defer func() {
		rsp["code"] = "00000"
		r, _ := json.Marshal(rsp)
		fmt.Fprintf(ctx, "%s", string(r))
	}()

	for _, p := range params {
		if p.Status != "0" {
			glog.Error("[mtsms] send fail, status：", p.Status)
			// 需要重发
			msg := &pb.RetrySendSms{Phone: p.Phone, Stype: data.MTSMS}
			nodePid.Tell(msg)
			continue
		}

		// 记录验证码
		log := &data.LogSmsRecord{
			Id:       bson.NewObjectId().String(),
			SmsId:    p.Uid,
			Reslut:   p.Status == "0",
			SmsCount: 1,
			Phone:    p.Phone,
			State:    p.Status,
			Channel:  data.MTSMS,
		}
		sendTime, _ := utils.Str2Unix(p.ReportTime)
		log.SendTime = sendTime
		log.Save()
	}
}

// xxy验证码回调
func xxySms(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "failure")
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
	glog.Infof("xxysms clientIP: %s", clientIP)
	//解析
	body := ctx.PostBody()
	rsp := make(map[string]string)
	// glog.Infof("paynotify:%s", string(body))
	params, err := smsxxy.ParseCallback(body)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}

	defer func() {
		rsp["code"] = "00000"
		r, _ := json.Marshal(rsp)
		fmt.Fprintf(ctx, "%s", string(r))
	}()

	for _, p := range params {
		if p.Status != "0" {
			glog.Error("[xxysms] send fail, status：", p.Status)
			// 需要重发
			msg := &pb.RetrySendSms{Phone: p.Phone, Stype: data.XXYSMS}
			nodePid.Tell(msg)
			continue
		}

		// 记录验证码
		log := &data.LogSmsRecord{
			Id:       bson.NewObjectId().String(),
			SmsId:    p.Uid,
			Reslut:   p.Status == "0",
			SmsCount: 1,
			Phone:    p.Phone,
			State:    p.Status,
			Channel:  data.XXYSMS,
		}
		sendTime, _ := utils.Str2Unix(p.ReportTime)
		log.SendTime = sendTime
		log.Save()
	}
}

// 创蓝验证码回调
func clSms(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "failure")
		}
	}()
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("CLSms clientIP: %s", clientIP)
	//解析
	body := ctx.QueryArgs().QueryString()
	// rsp := make(map[string]string)
	glog.Debug("body:", string(body))
	param, err := smscl.ParseCallback(body)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	// 记录验证码
	log := &data.LogSmsRecord{
		Id:       bson.NewObjectId().String(),
		SmsId:    param.Msgid,
		SendTime: utils.BsonNow().Unix(),
		Reslut:   param.Status == "DELIVRD",
		SmsCount: int(utils.Int64(param.SmsNum)),
		Phone:    param.Mobile,
		State:    param.Status,
		Channel:  data.CLSMS,
	}
	log.Save()

	// 判断是否发送成功
	if param.Status == "DELIVRD" {
		return
	}
	// 换渠道发
	glog.Error("clsms send fail, status：", param.Status)
	// 需要重发
	msg := &pb.RetrySendSms{Phone: param.Mobile[2:], Stype: data.CLSMS}
	nodePid.Tell(msg)
}

func kmiMjlSms(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			fmt.Fprintf(ctx, "%s", "failure")
		}
	}()
	ctx.Response.Header.Set("Content-type", "application/json")
	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	clientIP := getIP(ctx)
	glog.Infof("kmiMjlSms clientIP: %s", clientIP)
	//解析
	body := ctx.QueryArgs().QueryString()
	// rsp := make(map[string]string)
	glog.Debug("body:", string(body))
	param, err := smskmi.ParseCallback(body)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "failure")
		return
	}
	// 记录验证码
	log := &data.LogSmsRecord{
		Id:       bson.NewObjectId().String(),
		SmsId:    param.SmsId,
		SendTime: utils.Int64(param.SendTime),
		Reslut:   param.SendResult == "SUCCESS",
		SmsCount: 1,
		Phone:    param.To,
		State:    param.Dr,
		Channel:  data.KMIMJLSMS,
	}
	log.Save()

	// 判断是否发送成功
	if param.SendResult == "SUCCESS" {
		return
	}
	// 换渠道发
	glog.Errorf("kmiMjlSms send fail, status：%s,phone:%s", param.Dr, param.To)
	// 需要重发
	msg := &pb.RetrySendSms{Phone: param.To[3:], Stype: data.KMIMJLSMS}
	nodePid.Tell(msg)
}
