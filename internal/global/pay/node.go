package pay

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"runtime/debug"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/gogo/protobuf/proto"
)

var (
	nodePid *actor.PID
)

// PayActor 支付服务
type PayActor struct {
	Name string
	//关闭通道
	stopCh chan struct{}
	//计时
	timer int
}

// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
func (a *PayActor) Receive(ctx actor.Context) {
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

func newPayActor() actor.Actor {
	a := new(PayActor)
	//roles key=userid - "1"
	a.stopCh = make(chan struct{})
	return a
}

// NewRemote 启动服务
func NewRemote(bind, kind string) {
	remote.Start(bind)
	payProps := actor.FromProducer(newPayActor)
	remote.Register(kind, payProps)
	nodePid, err = actor.SpawnNamed(payProps, kind)
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
