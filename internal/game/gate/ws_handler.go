package gate

import (
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (ws *WSConn) SetLogined(ctx actor.Context) {
	//设置连接,还未登录
	arg := ctx.Message().(*pb.SetLogined)
	glog.Debugf("ws SetLogined %#v", arg)
	ws.rolePid = arg.RolePid
}

func (ws *WSConn) start(ctx actor.Context) {
	glog.Infof("ws start: %v", ctx.Self().String())
	ctx.SetReceiveTimeout(waitForLogin) //login timeout set
	msg2 := &pb.SetLogin{
		Sender: ctx.Self(),
	}
	//nodePid.Tell(msg2)
	timeout := 3 * time.Second
	res1, err1 := nodePid.RequestFuture(msg2, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRole err: %v", err1)
		return
	}
	if response1, ok := res1.(*pb.SetLogined); ok {
		ws.rolePid = response1.RolePid
	}
}

func (ws *WSConn) ServeClose(ctx actor.Context) {
	msg := ctx.Message().(*pb.ServeClose)
	glog.Debugf("ws ServeClose %#v", msg)
	glog.Infof("ws stop: %v", ctx.Self().String())
	//断开连接
	ws.Close()
	//表示已经断开
	ws.online = false
	ctx.Self().Stop()
}

func (ws *WSConn) ServeStop(ctx actor.Context) {
	msg := ctx.Message().(*pb.ServeStop)
	glog.Debugf("ws ServeStop %#v", msg)
	//断开连接
	// ws.ServeClose(ctx)
	// ctx.Self().Tell(new(pb.ServeClose))
	ws.pid = nil
	ws.Close()
	//表示已经断开
	ws.online = false
	ctx.Self().Stop()
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

func (ws *WSConn) ServeStart(ctx actor.Context) {
	msg := ctx.Message().(*pb.ServeStart)
	glog.Debugf("ws ServeStart %#v", msg)
	ws.start(ctx)
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)
}

func (ws *WSConn) AdjustAdidReq(ctx actor.Context) {
	// 还未登录完成就收到客户端的adid
	glog.Errorf("send adid error, user not login")
	rsp := new(pb.AdjustAdidRsp)
	rsp.Error = pb.SendLaterAdid
	ws.Send(rsp)
}
