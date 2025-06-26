package logger

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"runtime/debug"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/mailbox"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/AsynkronIT/protoactor-go/router"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	loggerPid *actor.PID
)

const maxConcurrency = 5

// LoggerActor 日志记录服务
type LoggerActor struct {
	Name string
}

// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
func (a *LoggerActor) Receive(ctx actor.Context) {
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
		a.Handler(msg, ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func newLoggerActor() actor.Actor {
	a := new(LoggerActor)
	//name
	a.Name = cfg.Section("logger").Name()
	return a
}

//func NewLogger() *actor.PID {
//	props := actor.FromProducer(newLoggerActor)
//	pid := actor.Spawn(props)
//	return pid
//}

// NewLoggerProps 启动
func NewLoggerProps() *actor.Props {
	return router.NewRoundRobinPool(maxConcurrency).
		WithProducer(newLoggerActor).
		WithMailbox(mailbox.Unbounded())
}

// NewRemote 启动服务
func NewRemote(bind, kind string) {
	remote.Start(bind, remote.WithServerOptions(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10M
		grpc.MaxSendMsgSize(10*1024*1024), // 10M
	))
	loggerProps := NewLoggerProps()
	remote.Register(kind, loggerProps)
	loggerPid, err = actor.SpawnNamed(loggerProps, kind)
	if err != nil {
		glog.Fatalf("loggerPid err %v", err)
	}
	glog.Infof("loggerPid %s", loggerPid.String())
}

// Stop 关闭服务
func Stop() {
	timeout := 5 * time.Second
	msg := new(pb.ServeStop)

	if loggerPid != nil {
		res1, err1 := loggerPid.RequestFuture(msg, timeout).Result()
		if err1 != nil {
			glog.Errorf("loggerPid Stop err: %v", err1)
		}
		if response1, ok := res1.(*pb.ServeStoped); ok {
			glog.Debugf("response1: %#v", response1)
		}
		loggerPid.Stop()
	}
}
