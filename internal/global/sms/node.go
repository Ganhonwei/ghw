package sms

import (
	"runtime/debug"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/gogo/protobuf/proto"
)

var (
	nodePid *actor.PID
)

// SmsActor 验证码服务
type SmsActor struct {
	Name string
	//中心服务
	// dbmsPid *actor.PID
	//发送电话 phone - code
	phoneCode map[string]string
	//发送电话 phone - overtime
	timeCode map[string]int64
	//关闭通道
	stopCh chan struct{}
	//更新状态
	status bool
	//计时
	timer int
}

// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
func (a *SmsActor) Receive(ctx actor.Context) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("Receive handler recover error:", r)
			debug.PrintStack()
		}
	}()
	switch msg := ctx.Message().(type) {
	case *pb.Request:
		ctx.Respond(&pb.Response{})
	case *actor.Started:
		glog.Notice("Starting, initialize actor here")
	case *actor.Stopping:
		glog.Notice("Stopping, actor is about to shut down")
	case *actor.Stopped:
		glog.Notice("Stopped, actor and its children are stopped")
	case *actor.Restarting:
		glog.Notice("Restarting, actor is about to restart")
	case *actor.ReceiveTimeout:
		glog.Infof("ReceiveTimeout: %v", ctx.Self().String())
	case proto.Message:
		a.Handler(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func newSmsActor() actor.Actor {
	a := new(SmsActor)
	a.Name = cfg.Section("sms").Name()
	//roles key=userid - "1"
	a.phoneCode = make(map[string]string)
	a.timeCode = make(map[string]int64)
	a.stopCh = make(chan struct{})
	return a
}

// NewRemote 启动服务
func NewRemote(bind, kind string) {
	remote.Start(bind)
	loginProps := actor.FromProducer(newSmsActor)
	remote.Register(kind, loginProps)
	nodePid, err = actor.SpawnNamed(loginProps, kind)
	if err != nil {
		glog.Fatalf("nodePid err %v", err)
	}
	glog.Infof("nodePid %s", nodePid.String())
	nodePid.Tell(new(pb.ServeStart))
}

// Stop 关闭服务
func Stop() {
	timeout := 5 * time.Second
	msg := new(pb.ServeStop)
	if nodePid != nil {
		res1, err1 := nodePid.RequestFuture(msg, timeout).Result()
		if err1 != nil {
			glog.Errorf("nodePid Stop err: %v", err1)
		}
		response1 := res1.(*pb.ServeStoped)
		glog.Debugf("response1: %#v", response1)
		nodePid.Stop()
	}
}
