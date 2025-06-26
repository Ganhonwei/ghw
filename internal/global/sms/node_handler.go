package sms

import (
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 启动服务
func (a *SmsActor) ServeStart(ctx actor.Context) {
	glog.Infof("desk start: %v", ctx.Self().String())
	//dbms
	// bind := cfg.Section("dbms").Key("bind").Value()
	// kind := cfg.Section("dbms").Key("kind").Value()
	// a.dbmsPid = actor.NewPID(bind, kind)
	// glog.Infof("a.dbmsPid: %s", a.dbmsPid.String())
	// connect := &pb.Connect{
	// 	Name: a.Name,
	// }
	// a.dbmsPid.Request(connect, ctx.Self())
	//启动
	go a.ticker(ctx)
	//响应
	rsp := new(pb.ServeStarted)
	ctx.Respond(rsp)
}

// 停止服务
func (a *SmsActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//TODO 关闭消息
	//断开处理
	// msg := &pb.Disconnect{
	// 	Name: a.Name,
	// }
	// if a.dbmsPid != nil {
	// 	a.dbmsPid.Tell(msg)
	// }
	//延迟
	<-time.After(2 * time.Second)
	//响应
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

// 滴答
func (a *SmsActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//TODO 同步节点
	//a.dbmsPid.Request(msg, ctx.Self())
	//TODO 路由方式
	//err := cfg.Section("gate.node"+version).NewKey("host", "x:x")
	a.expireCode()
}

// 时钟
func (a *SmsActor) ticker(ctx actor.Context) {
	tick := time.Tick(5 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

func (a *SmsActor) RetrySendSms(ctx actor.Context) {
	arg := ctx.Message().(*pb.RetrySendSms)
	glog.Debugf("RetrySendSms %#v", arg)
	if code, ok := a.phoneCode[arg.Phone]; ok {
		a.timeCode[arg.Phone] = utils.BsonNow().Unix() + 120 // 保存2分钟
		go a.sendSms(arg.Phone, code, int(arg.Stype+1))
	}
}

// 关闭时钟
func (a *SmsActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *SmsActor) SendSmsCode(ctx actor.Context) {
	arg := ctx.Message().(*pb.SendSmsCode)
	glog.Debugf("SendSmsCode %#v", arg)
	a.phoneCode[arg.Phone] = arg.Code
	a.timeCode[arg.Phone] = utils.BsonNow().Unix() + 120 // 保存2分钟
	s := data.KMISMS
	if strings.HasPrefix(arg.Phone, "880") || strings.HasPrefix(arg.Phone, "92") {
		s = data.KMIMJLSMS
	}
	go a.sendSms(arg.Phone, arg.Code, s)
}

func (a *SmsActor) sendSms(phone, code string, n int) {
	// if cfg.Section("smsbk").Key("status").MustBool(false) {
	// 	go login.SendSmsBK(smsbkURL, smsbkappid, arg.Phone, code, smsbkaccount, smsbkpassword)
	// } else
	// a.phoneCode[phone] = code
	// a.timeCode[phone] = utils.BsonNow().Unix() + 180 // 保存3分钟
	success := false
	switch n {
	case 0:
		if cfg.Section("smskmi").Key("status").MustBool(false) {
			success = login.SendSmsKMI(SmskmiURL, SmskmiAccKey, SmskmiSecKey, SmskmiSender, SmskmiTemp, phone, code)
		}
	case 1:
		if cfg.Section("smsmt").Key("status").MustBool(false) {
			success = login.SendSmsMT(SmsmtURL, SmsmtAppKey, SmsmtAppSecret, SmsmtAppCode, SmsmtTemp, phone, code)
		}
	case 2:
		if cfg.Section("smscl").Key("status").MustBool(false) {
			success = login.SendSmsCL(SmsclUrl, SmsclApiKey, SmsclPassword, phone, code)
		}
	case 3:
		if cfg.Section("smsxxy").Key("status").MustBool(false) {
			success = login.SendSmsXXY(SmsxxyURL, SmsxxyAppKey, SmsxxyAppSecret, SmsxxyAppCode, SmsxxyTemp, phone, code)
		}
	case 4:
		if cfg.Section("smsmjlkmi").Key("status").MustBool(false) {
			success = login.SendSmsKMI(SmsMJLkmiURL, SmsMJLkmiAccKey, SmsMJLkmiSecKey, SmsMJLkmiSender, SmsMJLkmiTemp, phone, code)
		}
	default:
		glog.Errorf("send smscode fail,phone:%s,code:%s", phone, code)
		return
	}
	if success {
		return
	}
	a.sendSms(phone, code, n+1)
}

// 过期验证码
func (a *SmsActor) expireCode() {
	now := utils.BsonNow().Unix()
	for k, v := range a.timeCode {
		if now > v {
			delete(a.timeCode, k)
			delete(a.phoneCode, k)
		}
	}
}
