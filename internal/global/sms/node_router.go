package sms

import (
	"goserver/gen/pb"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *SmsActor) Handler(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.Ping:
		ctx.Respond(new(pb.Pong))
	case *pb.SendSmsCode:
		a.SendSmsCode(ctx)
	case *pb.RetrySendSms:
		a.RetrySendSms(ctx)
	default:
		// glog.Errorf("unknown message %v", msg)
	}
}
