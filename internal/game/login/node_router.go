package login

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *LoginActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.Tick:
		a.Tick(ctx)
	case *pb.WxpayCallback:
		a.WxpayCallback(ctx)
	case *pb.SmscodeRegist:
		a.SmscodeRegist(ctx)
	case *pb.WebRequest:
		a.WebRequest(ctx)
	case *pb.TradeOrder:
		a.TradeOrder(ctx)
	case *pb.JtpayCallback:
		a.JtpayCallback(ctx)
	// case *pb.AgentConfirm:
	// 	a.AgentConfirm(ctx)
	// case *pb.AgentOauth2Confirm:
	// 	a.AgentOauth2Confirm(ctx)
	case *pb.SyncConfig:
		a.SyncConfig(ctx)
	case *pb.PayOrderUpdate:
		a.PayOrderUpdate(ctx)
	case *pb.LogRechargeReport:
		a.LogRechargeReport(ctx)
	case *pb.LogSmsRecord:
		a.LogSmsRecord(ctx)
	case *pb.GetUserInfoReq:
		a.GetUserInfoReq(ctx)
	case *pb.ExternalBetReq:
		a.ExternalBetReq(ctx)
	case *pb.ExternalRewardReq:
		a.ExternalRewardReq(ctx)
	case *pb.ExternalCancelReq:
		a.ExternalCancelReq(ctx)
	case *pb.EfiTransactionReq:
		a.EfiTransactionReq(ctx)
	case *pb.RepeatWithdraw:
		a.RepeatWithdraw(ctx)
	case *pb.WebPageAdid:
		a.WebPageAdid(ctx)
	case *pb.PlayshareTest:
		a.PlayshareTest(ctx)
	case *pb.UploadPhotoLog:
		a.UploadPhotoLog(ctx)
	case *pb.UploadUtrReq:
		a.UploadUtrReq(ctx)
	case *pb.CustomerUploadFile:
		a.CustomerUploadFile(ctx)
	case *pb.Ping:
		a.Ping(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
