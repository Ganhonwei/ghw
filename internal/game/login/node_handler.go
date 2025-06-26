package login

import (
	"fmt"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 启动服务
func (a *LoginActor) ServeStart(ctx actor.Context) {
	glog.Infof("desk start: %v", ctx.Self().String())
	myactor.InitLogger(cfg)
	//dbms
	bind := cfg.Section("dbms").Key("bind").Value()
	kind := cfg.Section("dbms").Key("kind").Value()
	role := cfg.Section("dbms").Key("role").Value()
	room := cfg.Section("dbms").Key("room").Value()
	// logger := cfg.Section("dbms").Key("logger").Value()
	a.dbmsPid = actor.NewPID(bind, kind)
	a.rolePid = actor.NewPID(bind, role)
	a.roomPid = actor.NewPID(bind, room)
	// a.loggerPid = actor.NewPID(bind, logger)
	glog.Infof("a.dbmsPid: %s", a.dbmsPid.String())
	glog.Infof("a.rolePid: %s", a.rolePid.String())
	glog.Infof("a.roomPid: %s", a.rolePid.String())
	connect := &pb.Connect{
		Name: a.Name,
	}
	a.dbmsPid.Request(connect, ctx.Self())
	//启动
	go a.ticker(ctx)
	//响应
	rsp := new(pb.ServeStarted)
	ctx.Respond(rsp)
}

// 停止服务
func (a *LoginActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//TODO 关闭消息
	//断开处理
	msg := &pb.Disconnect{
		Name: a.Name,
	}
	if a.dbmsPid != nil {
		a.dbmsPid.Tell(msg)
	}
	//延迟
	<-time.After(2 * time.Second)
	//响应
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

// 滴答
func (a *LoginActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//TODO 同步节点
	//a.dbmsPid.Request(msg, ctx.Self())
	//TODO 路由方式
	//err := cfg.Section("gate.node"+version).NewKey("host", "x:x")
	switch a.timer {
	case 2: //1分钟
		a.smsExpire()
		a.timer = 0
	default:
		a.timer++
	}
}

// 时钟
func (a *LoginActor) ticker(ctx actor.Context) {
	tick := time.Tick(30 * time.Second)
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

func (a *LoginActor) WxpayCallback(ctx actor.Context) {
	arg := ctx.Message().(*pb.WxpayCallback)
	glog.Debugf("WxpayCallback: %v", arg)
	a.rolePid.Tell(arg)
}

func (a *LoginActor) SmscodeRegist(ctx actor.Context) {
	arg := ctx.Message().(*pb.SmscodeRegist)
	glog.Debugf("SmscodeRegist %v", arg)
	//验证限制
	if v, ok := a.smsphone[arg.Phone]; ok && v > 10 {
		glog.Debugf("SmscodeRegist Phone %s, v %d", arg.Phone, v)
		return
	}
	if v, ok := a.smstimes[arg.Ipaddr]; ok && v > 10 {
		glog.Debugf("SmscodeRegist Ipaddr %s, v %d", arg.Ipaddr, v)
		return
	}
	a.smsphone[arg.Phone]++  //次数
	a.smstimes[arg.Ipaddr]++ //3分钟
	a.rolePid.Tell(arg)
}

func (a *LoginActor) WebRequest(ctx actor.Context) {
	arg := ctx.Message().(*pb.WebRequest)
	glog.Debugf("WebRequest %#v", arg)
	var res1 interface{}
	var err1 error
	switch arg.Code {
	case pb.WebOnline, pb.WebBuild,
		pb.WebGive, pb.WebNumber, pb.WebGiveWithdraw,
		pb.WebRate, pb.WebState, pb.WebOnlineUser,
		pb.WebVaild, pb.WebModifyUser, pb.WebPointControl,
		pb.WebBlack, pb.WebWithdraw, pb.WebSuperior,
		pb.WebModifyNum, pb.WebFeedBack, pb.WebPayCallback, pb.WebTpStoryStockMin,
		pb.WebUserPhoto, pb.WebTurnAudit:
		res1, err1 = a.callRole(arg)
	case pb.WebModifyStock, pb.WebNewbieStock, pb.WebModifyTpStoryStock, pb.WebGameStock:
		res1, err1 = a.callRoom(arg)
	default:
		res1, err1 = a.callDbms(arg)
	}
	if err1 != nil {
		rsp := new(pb.WebResponse)
		rsp.ErrMsg = fmt.Sprintf("dbms request err1 %v", err1)
		ctx.Respond(rsp)
		return
	}
	if res2, ok := res1.(*pb.WebResponse); !ok {
		rsp := new(pb.WebResponse)
		rsp.ErrMsg = fmt.Sprintf("dbms response err %#v", res2)
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res1)
}

func (a *LoginActor) TradeOrder(ctx actor.Context) {
	arg := ctx.Message().(*pb.TradeOrder)
	glog.Debugf("TradeOrder %#v", arg)
	res1, err1 := a.callRole(arg)
	if err1 != nil {
		glog.Errorf("TradeOrder response err %v", err1)
		rsp := new(pb.TradedOrder)
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res1)
}

func (a *LoginActor) JtpayCallback(ctx actor.Context) {
	arg := ctx.Message().(*pb.JtpayCallback)
	glog.Debugf("JtpayCallback: %v", arg)
	res1, err1 := a.callRole(arg)
	if err1 != nil {
		glog.Errorf("JtpayCallback response err %v", err1)
		rsp := new(pb.JtpayCalledback)
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res1)
}

// func (a *LoginActor) AgentConfirm(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.AgentConfirm)
// 	glog.Debugf("AgentConfirm: %v", arg)
// 	res1, err1 := a.callRole(arg)
// 	if err1 != nil {
// 		glog.Errorf("AgentConfirm response err %v", err1)
// 		rsp := new(pb.AgentConfirmed)
// 		rsp.Error = pb.Failed
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	ctx.Respond(res1)
// }

// func (a *LoginActor) AgentOauth2Confirm(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.AgentOauth2Confirm)
// 	glog.Debugf("AgentOauth2Confirm: %v", arg)
// 	res1, err1 := a.callDbms(arg)
// 	if err1 != nil {
// 		glog.Errorf("AgentOauth2Confirm response err %v", err1)
// 		rsp := new(pb.AgentOauth2Confirmed)
// 		rsp.Error = pb.Failed
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	ctx.Respond(res1)
// }

func (a *LoginActor) SyncConfig(ctx actor.Context) {
	arg := ctx.Message().(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	handler.SyncConfig(arg, a.Name)
}

func (a *LoginActor) PayOrderUpdate(ctx actor.Context) {
	arg := ctx.Message().(*pb.PayOrderUpdate)
	glog.Debugf("PayOrderUpdate %#v", arg)
	res1, err1 := a.callRole(arg)
	if err1 != nil {
		glog.Errorf("payCallback response err %v", err1)
		rsp := new(pb.PayedOrderUpdate)
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res1)
}

func (a *LoginActor) RepeatWithdraw(ctx actor.Context) {
	arg := ctx.Message().(*pb.RepeatWithdraw)
	glog.Debugf("RepeatWithdraw %#v", arg)
	res1, err1 := a.callRole(arg)
	if err1 != nil {
		glog.Errorf("RepeatWithdraw response err %v", err1)
		rsp := new(pb.RepeatWithdrawed)
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res1)
}

func (a *LoginActor) LogRechargeReport(ctx actor.Context) {
	arg := ctx.Message().(*pb.LogRechargeReport)
	glog.Debugf("LogRechargeReport %#v", arg)
	myactor.Logger().Tell(arg)
}

func (a *LoginActor) LogSmsRecord(ctx actor.Context) {
	arg := ctx.Message().(*pb.LogSmsRecord)
	glog.Debugf("LogSmsRecord %#v", arg)
	myactor.Logger().Tell(arg)
}

func (a *LoginActor) GetUserInfoReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.GetUserInfoReq)
	glog.Debugf("GetUserInfoReq %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.GetUserInfoRsp)
		rsp.Err = err.Error()
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) ExternalBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalBetReq)
	glog.Debugf("ExternalBetReq %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.ExternalBetRsp)
		rsp.Err = err.Error()
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) ExternalRewardReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalRewardReq)
	glog.Debugf("ExternalRewardReq %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.ExternalRewardRsp)
		rsp.Err = err.Error()
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) ExternalCancelReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ExternalCancelReq)
	glog.Debugf("ExternalCancelReq %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.ExternalCancelRsp)
		rsp.Err = err.Error()
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) EfiTransactionReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.EfiTransactionReq)
	glog.Debugf("EfiTransactionReq %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.EfiTransactionRsp)
		rsp.Err = err.Error()
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) WebPageAdid(ctx actor.Context) {
	arg := ctx.Message().(*pb.WebPageAdid)
	glog.Debugf("WebPageAdid %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		rsp := new(pb.WebPageAdided)
		rsp.Code = -1
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) PlayshareTest(ctx actor.Context) {
	arg := ctx.Message().(*pb.PlayshareTest)
	glog.Debugf("PlayshareTest %#v", arg)
	res, err := a.callRole(arg)
	if err != nil {
		ctx.Respond("")
		return
	}
	ctx.Respond(res)
}

func (a *LoginActor) UploadPhotoLog(ctx actor.Context) {
	arg := ctx.Message().(*pb.UploadPhotoLog)
	glog.Debugf("UploadPhotoLog %#v", arg)
	myactor.Logger().Tell(arg)
	a.rolePid.Tell(arg)
}

func (a *LoginActor) UploadUtrReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UploadUtrReq)
	glog.Debugf("UploadUtrReq %#v", arg)
	a.forwardRole(arg)
}

func (a *LoginActor) CustomerUploadFile(ctx actor.Context) {
	arg := ctx.Message().(*pb.CustomerUploadFile)
	glog.Debugf("CustomerUploadFile %#v", arg)
	a.forwardRole(arg)
}

// func (a *LoginActor) RetrySendSms(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.RetrySendSms)
// 	glog.Debugf("RetrySendSms %#v", arg)
// 	res1, err1 := a.callRole(arg)
// 	if err1 != nil {
// 		glog.Errorf("RetrySendSms response err %v", err1)
// 		rsp := new(pb.RetrySendSmsed)
// 		ctx.Respond(rsp)
// 		return
// 	}
// 	ctx.Respond(res1)
// }

func (a *LoginActor) Ping(ctx actor.Context) {
	res, err := a.callDbms(ctx.Message())
	if err != nil {
		rsp := &pb.Pong{
			Msg: "failed",
		}
		// rsp.Msg = fmt.Sprintf("dbms request err1 %v", err1)
		ctx.Respond(rsp)
		return
	}
	if _, ok := res.(*pb.Pong); !ok {
		rsp := &pb.Pong{
			Msg: "failed",
		}
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(res)
}

// 过期检测
func (a *LoginActor) smsExpire() {
	for k, v := range a.smsphone {
		if v >= 1 {
			a.smsphone[k]--
			continue
		}
		delete(a.smsphone, k)
	}
	for k, v := range a.smstimes {
		if v >= 1 {
			a.smstimes[k]--
			continue
		}
		delete(a.smstimes, k)
	}
}

// 关闭时钟
func (a *LoginActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *LoginActor) callDbms(msg interface{}) (interface{}, error) {
	timeout := 5 * time.Second
	res, err := a.dbmsPid.RequestFuture(msg, timeout).Result()
	return res, err
}

func (a *LoginActor) callRole(msg interface{}) (interface{}, error) {
	timeout := 5 * time.Second
	res, err := a.rolePid.RequestFuture(msg, timeout).Result()
	return res, err
}

func (a *LoginActor) forwardRole(msg interface{}) {
	a.rolePid.Tell(msg)
}

func (a *LoginActor) callRoom(msg interface{}) (interface{}, error) {
	timeout := 5 * time.Second
	res, err := a.roomPid.RequestFuture(msg, timeout).Result()
	return res, err
}

func callNode(msg interface{}) (interface{}, error) {
	timeout := 5 * time.Second
	res, err := nodePid.RequestFuture(msg, timeout).Result()
	return res, err
}

func forwardNode(msg interface{}) {
	nodePid.Tell(msg)
}
