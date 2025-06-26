package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (ws *WSConn) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.SetLogined:
		ws.SetLogined(ctx)
	case *pb.ServeClose:
		ws.ServeClose(ctx)
	case *pb.ServeStop:
		ws.ServeStop(ctx)
	case *pb.ServeStoped:
	case *pb.ServeStart:
		ws.ServeStart(ctx)
	case *pb.ServeStarted:
	case proto.Message:
		//响应消息
		//ws.Send(msg)
		ws.handlerLogin(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家数据请求处理
func (ws *WSConn) handlerLogin(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.PingReq:
		ws.PingReq(ctx)
	case *pb.RegistReq:
		ws.RegistReq(ctx)
	case *pb.LoginReq:
		ws.LoginReq(ctx)
	// case *pb.WxLoginReq:
	// 	ws.WxLoginReq(ctx)
	case *pb.ResetPwdReq:
		//TODO 暂时屏蔽使用
		// ws.ResetPwdReq(ctx)
	case *pb.TouristReq:
		ws.TouristReq(ctx)
	case *pb.ButtonClickReq:
	case *pb.AdjustAdidReq:
		ws.AdjustAdidReq(ctx)
	case *pb.LoginElse:
		ws.LoginElse(ctx)
	case proto.Message:
		//响应
		if ws.online {
			ws.Send(msg)
		}
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
