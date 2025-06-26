package logger

import (
	"context"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// Handler 消息处理
func (a *LoggerActor) Handler(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.LogRegist:
		arg := msg.(*pb.LogRegist)
		data.RegistRecord(arg.Userid, arg.Nickname, arg.Ip, arg.Atype, location)
	case *pb.LogLogin:
		arg := msg.(*pb.LogLogin)
		data.LoginRecord(arg.Userid, arg.Ip, arg.Atype, location)
	case *pb.LogLogout:
		arg := msg.(*pb.LogLogout)
		data.LogoutRecord(arg.Userid, int(arg.Event))
	case *pb.LogDiamond:
		arg := msg.(*pb.LogDiamond)
		data.DiamondRecord(arg.Userid, arg.Type, arg.Rest, arg.Num)
	case *pb.LogCoin:
		arg := msg.(*pb.LogCoin)
		data.CoinRecord(arg.Userid, arg.Type, arg.Rest, arg.Num)
	case *pb.LogCard:
		arg := msg.(*pb.LogCard)
		data.CardRecord(arg.Userid, arg.Type, arg.Rest, arg.Num)
	case *pb.LogChip:
		arg := msg.(*pb.LogChip)
		data.ChipRecord(arg.Userid, arg.Type, arg.Rest, arg.Num)
	case *pb.LogTask:
		arg := msg.(*pb.LogTask)
		data.TaskRecord(arg.Userid, arg.Taskid, arg.Type)
	// case *pb.LogNotice:
	// 	arg := msg.(*pb.LogNotice)
	// 	handler.SaveNotice(arg)
	case *pb.LogBank:
		arg := msg.(*pb.LogBank)
		data.BankRecord(arg.Userid, arg.From, arg.Type, arg.Rest, arg.Num)
	case *pb.RoomRecordInfo:
		arg := msg.(*pb.RoomRecordInfo)
		handler.Log2RoomRecord(arg)
	case *pb.RoleRecord:
		arg := msg.(*pb.RoleRecord)
		handler.Log2RoleRecord(arg)
	case *pb.RoundRecord:
		arg := msg.(*pb.RoundRecord)
		handler.Log2RoundRecord(arg)
	case *pb.Detail:
		arg := msg.(*pb.Detail)
		handler.Log2Detail(arg)
	case *pb.LogProfit:
		arg := msg.(*pb.LogProfit)
		data.ProfitRecord(arg)
	case *pb.LogSysProfit:
		arg := msg.(*pb.LogSysProfit)
		data.SysProfitRecord(arg.Agentid, arg.Userid, arg.Gtype,
			arg.Level, arg.Rate, arg.Profit, arg.Rest)
	case *pb.LogCurrencyWater:
		arg := msg.(*pb.LogCurrencyWater)
		data.CurrencyRecord(arg)
	case *pb.LogWithdrawCash:
		arg := msg.(*pb.LogWithdrawCash)
		data.OutDiamondLog(arg)
	case *pb.LogShareWithDraw:
		arg := msg.(*pb.LogShareWithDraw)
		data.LogShare(arg)
	case *pb.LogEventTrack:
		arg := msg.(*pb.LogEventTrack)
		data.EventTrack(arg)
	case *pb.LogBugFeedback:
		arg := msg.(*pb.LogBugFeedback)
		data.BugLog(arg)
	case *pb.LogGameTime:
		arg := msg.(*pb.LogGameTime)
		data.GameTime(arg)
	case *pb.LogButtonClick:
		arg := msg.(*pb.LogButtonClick)
		data.ButtonClick(arg)
	case *pb.LogButtonClick2:
		arg := msg.(*pb.LogButtonClick2)
		data.ButtonClick2(arg)
	case *pb.LogRechargeReport:
		arg := msg.(*pb.LogRechargeReport)
		data.RechargeReport(arg)
	case *pb.FeedBackReq:
		arg := msg.(*pb.FeedBackReq)
		data.FeedBackReq(arg)
	case *pb.LogSmsRecord:
		arg := msg.(*pb.LogSmsRecord)
		data.LogSmsRecordReq(arg)
	case *pb.LogUserSmsRecord:
		arg := msg.(*pb.LogUserSmsRecord)
		data.LogUserSmsRecord(arg)
	case *pb.ScratchTicketLog:
		arg := msg.(*pb.ScratchTicketLog)
		data.LogScratchTicketData(arg)
	case *pb.ActivityMonitorReq:
		arg := msg.(*pb.ActivityMonitorReq)
		data.ActivityMonitorLog(arg)
	case *pb.PlayShareDrawLog:
		arg := msg.(*pb.PlayShareDrawLog)
		data.PlayShareDrawLog(arg)
	case *pb.BonusLog:
		arg := msg.(*pb.BonusLog)
		data.BonusLog(arg)
	case *pb.GameFlowWaterLog:
		arg := msg.(*pb.GameFlowWaterLog)
		data.GameFlowWaterLog(arg)
	case *pb.UploadPhotoLog:
		arg := msg.(*pb.UploadPhotoLog)
		data.UploadUserHeadRecord(arg)
		myredis.Redis().Set(context.Background(), data.GetCustomPhotoKey(arg.Userid), arg.Url, 0)
	case *pb.ServeStop:
		//关闭服务
		a.handlerStop(ctx)
		//响应登录
		rsp := new(pb.ServeStoped)
		ctx.Respond(rsp)
	case *pb.JHVFRecordSync: // tp vf记录
		arg := msg.(*pb.JHVFRecordSync)
		data.LogJHVFRecordSync(arg)
	case *pb.LaunchInviteLog: // 点击邀请按钮日志
		arg := msg.(*pb.LaunchInviteLog)
		data.LogLaunchInvite(arg)
	case *pb.VolatilitySubsidyPool:
		arg := msg.(*pb.VolatilitySubsidyPool)
		data.LogVolatilitySubsidyPool(arg)
	case *pb.OnlineUsersLog:
		arg := msg.(*pb.OnlineUsersLog)
		data.LogOnlineUsersLog(arg)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func (a *LoggerActor) handlerStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//TODO clean mailbox
}
