package sms

import (
	"fmt"
	"goserver/pkg/glog"

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
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("fail request path:", ctx.Path())
			fmt.Fprintf(ctx, "%s", r)
		}
	}()
	defer ctx.SetConnectionClose()
	switch string(ctx.Path()) {
	case "/api/reloadconfig":
		SmsInit()
	case "/api/sendSMSCode":
		sendSMSCode(ctx)
	default:
		smsRequestHandler(ctx)
	}
}

// 短信回调
func smsRequestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/smsnotify/kmi":
		kmiSms(ctx)
	case "/api/smsnotify/kmimjl":
		kmiMjlSms(ctx)
	case "/api/smsnotify/mt":
		mtSms(ctx)
	case "/api/smsnotify/xxy":
		xxySms(ctx)
	case "/api/smsnotify/cl":
		clSms(ctx)
	default:
		glog.Errorf("send error api:%s, status:%s", string(ctx.Path()), fasthttp.StatusNotFound)
	}
}
