package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (a *GateActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.Connected:
		a.Connected(ctx)
	case *pb.Disconnected:
		a.Disconnected(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	case *pb.Ping:
		ctx.Respond(new(pb.Pong))
	case proto.Message:
		a.handlerUser(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func (a *GateActor) handlerUser(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.PayCurrency:
		a.PayCurrency(ctx)
	case *pb.ChangeCurrency:
		a.ChangeCurrency(ctx)
	case *pb.ModifyCurrency:
		a.ModifyCurrency(ctx)
	case *pb.WxpayCallback:
		a.WxpayCallback(ctx)
	case *pb.PayGoods:
		a.PayGoods(ctx)
	case *pb.MarQueeNtf:
		a.MarQueeNtf(ctx)
	// case *pb.OnlineUser:
	// 	a.OnlineUser(ctx)
	case *pb.GamePeopleNtf:
		a.GamePeopleNtf(ctx)
	case *pb.PublishActivityBetRankUpdate:
		a.PublishActivityBetRankUpdate(ctx)
	case *pb.PublishPayChannelUpdate:
		a.PublishPayChannelUpdate(ctx)
	case *pb.ConsumerReplyMessage:
		a.ConsumerReplyMessage(ctx)
	case *pb.PublishCustomerSessionOver:
		a.PublishCustomerSessionOver(ctx)
	case *pb.PublishCustomerMessageRevoke:
		a.PublishCustomerMessageRevoke(ctx)
	case proto.Message:
		a.handlerLogin(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家数据请求处理
func (a *GateActor) handlerLogin(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.LoginElse:
		a.loginElse(ctx)
	case *pb.LoginedElse:
		a.LoginedElse(ctx)
	case *pb.SetLogin:
		a.SetLogin(ctx)
	case *pb.Logout:
		a.Logout(ctx)
	case *pb.SelectGate:
		a.SelectGate(ctx)
	//case *pb.Login2Gate:
	//	//登录成功
	//	arg := msg.(*pb.Login2Gate)
	//	glog.Debugf("Login2Gate %#v", arg)
	//	a.spawnRole(arg, ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
