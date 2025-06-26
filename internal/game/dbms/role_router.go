package dbms

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (a *RoleActor) Handler(ctx actor.Context) {
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
	case *pb.Tick2:
		a.Tick2(ctx)
	case *pb.Tick3:
		a.Tick3(ctx)
	case *pb.CloseServer:
		a.CloseServer(ctx)
	case proto.Message:
		a.handlerUser(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家请求处理
func (a *RoleActor) handlerUser(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.GetUser:
		a.GetUser(ctx)
	case *pb.SyncUser:
		a.SyncUser(ctx)
	case *pb.ChangeCurrency:
		a.ChangeCurrency(ctx)
	case *pb.OfflineCurrency:
		a.OfflineCurrency(ctx)
	case *pb.PayCurrency:
		a.PayCurrency(ctx)
	case *pb.ModifyCurrency:
		a.ModifyCurrency(ctx)
	case *pb.ChangePointControl:
		a.ChangePointControl(ctx)
	case *pb.LoginElse:
		a.LoginElse(ctx)
	case *pb.Login:
		a.Login(ctx)
	case *pb.Logout:
		a.Logout(ctx)
	case *pb.RoleRegist:
		a.RoleRegist(ctx)
	case *pb.RoleLogin:
		a.RoleLogin(ctx)
	case *pb.TouristLogin:
		a.TouristLogin(ctx)
	case *pb.WxLogin:
		a.WxLogin(ctx)
	case *pb.GetUserData:
		a.GetUserData(ctx)
	case *pb.ApplePay:
		a.ApplePay(ctx)
	case *pb.WxpayCallback:
		a.WxpayCallback(ctx)
	case *pb.TradeOrder:
		a.TradeOrder(ctx)
	case *pb.SmscodeRegist:
		a.SmscodeRegist(ctx)
	case *pb.ResetPwdReq:
		a.ResetPwdReq(ctx)
	case *pb.GetNumber:
		a.GetNumber(ctx)
	case *pb.WebRequest:
		a.WebRequest(ctx)
	case *pb.BankGive:
		a.BankGive(ctx)
	case *pb.BankChange:
		a.BankChange(ctx)
	case *pb.TaskUpdate:
		a.TaskUpdate(ctx)
	case *pb.RobotRegist:
		a.RobotRegist(ctx)
	case *pb.PushNoticeNtf:
		a.PushNoticeNtf(ctx)
	case *pb.BindPhoneReq:
		a.BindPhoneReq(ctx)
	case *pb.PayOrder:
		a.PayOrder(ctx)
	case *pb.PayOrderUpdate:
		a.PayOrderUpdate(ctx)
	case *pb.WithdrawRecordRefundReq:
		a.WithdrawRecordRefundReq(ctx)
	case *pb.PaymentRecordUTRReq:
		a.PaymentRecordUTRReq(ctx)
	case *pb.WithdrawRecordKeepWaitingReq:
		a.WithdrawRecordKeepWaitingReq(ctx)
	case *pb.WithdrawRecordOvertimePayTackReq:
		a.WithdrawRecordOvertimePayTackReq(ctx)
	case *pb.RepeatWithdraw:
		a.RepeatWithdraw(ctx)
	case *pb.FeedBackReq:
		a.FeedBackReq(ctx)
	case *pb.ConsumerReplyMessage:
		a.ConsumerReplyMessage(ctx)
	case *pb.PublishCustomerSessionOver:
		a.PublishCustomerSessionOver(ctx)
	case *pb.MarQueeNtf:
		a.MarQueeNtf(ctx)
	case *pb.MarQueeWithdrawNtf:
		a.MarQueeWithdrawNtf(ctx)
	case *pb.ShareUserRegist:
		a.ShareUserRegist(ctx)
	case *pb.GiveAndOutCash:
		a.GiveAndOutCash(ctx)
	case *pb.BindLoginPhoneReq:
		a.BindLoginPhoneReq(ctx)
	case *pb.UpdateRegisReward:
		a.UpdateRegisReward(ctx)
	case *pb.ShareRecharge:
		a.ShareRecharge(ctx)
	case *pb.ShareRegist:
		a.ShareRegist(ctx)
	case *pb.GetUserInfoReq:
		a.GetUserInfoReq(ctx)
	case *pb.ExternalBetReq:
		a.ExternalBetReq(ctx)
	case *pb.ExternalRewardReq:
		a.ExternalRewardReq(ctx)
	case *pb.ExternalCancelReq:
		a.ExternalCancelReq(ctx)
	// case *pb.EfiTransactionReq:
	// 	a.EfiTransactionReq(ctx)
	case *pb.WebPageAdid:
		a.WebPageAdid(ctx)
	case *pb.CDKReq:
		a.CDKReq(ctx)
	case *pb.PlayshareTest:
		a.PlayshareTest(ctx)
	case *pb.UploadPhotoLog:
		a.UploadPhoto(ctx)
	case *pb.ChangeViceAccount:
		a.ChangeViceAccount(ctx)
	case *pb.ShareAgentRegist:
		a.ShareAgentRegist(ctx)
	case *pb.ShareAgentBets:
		a.ShareAgentBets(ctx)
	case *pb.ShareAgentPay:
		a.ShareAgentPay(ctx)
	case *pb.ShareOngoingEarningUnclaimedTackReq:
		a.ShareOngoingEarningUnclaimedTackReq(ctx)
	case *pb.ShareInstantEarningUnclaimedTackReq:
		a.ShareInstantEarningUnclaimedTackReq(ctx)
	case *pb.ShareAgentIncomeTackReq:
		a.ShareAgentIncomeTackReq(ctx)
	case *pb.ShareAgentZeroReset:
		a.ShareAgentZeroReset(ctx)
	case *pb.GiftPackCodeTackReq:
		a.GiftPackCodeTackReq(ctx)
	case *pb.VolatilitySubsidyCheck:
		a.VolatilitySubsidyCheck(ctx)
	case *pb.VolatilitySubsidyToLobby:
		a.VolatilitySubsidyToLobby(ctx)
	case *pb.VolatilitySubsidyWheelReq:
		a.VolatilitySubsidyWheelReq(ctx)
	case *pb.HasFailUtr:
		a.HasFailUtr(ctx)
	case *pb.WithdrawTransfer:
		a.WithdrawTransfer(ctx)
	case *pb.CustomerUploadFile:
		a.CustomerUploadFile(ctx)
	// case *pb.RetrySendSms:
	// a.RetrySendSms(ctx)
	case proto.Message:
		a.handlerActivity(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 活动
func (a *RoleActor) handlerActivity(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.FirstRecharged:
		a.FirstRecharged(ctx)
	case *pb.OnlineRewardRecord:
		a.OnlineRewardRecord(ctx)
	case *pb.WeeklyCardReceiveReq:
		a.WeeklyCardReceiveReq(ctx)
	case *pb.ShareSettlement:
		a.ShareSettlement(ctx)
	case *pb.BugFeedbackReq:
		a.BugFeedbackReq(ctx)
	case *pb.LuckyDrawNumber:
		a.LuckyDrawNumber(ctx)
	case *pb.LuckyDrawReq:
		a.LuckyDrawReq(ctx)
	case *pb.LuckyDrawHistoryReq:
		a.LuckyDrawHistoryReq(ctx)
	case *pb.LuckyDrawTake:
		a.LuckyDrawTake(ctx)
	case *pb.ActivityBetRankWindowNtfs:
		a.ActivityBetRankWindowNtfs(ctx)
	case *pb.ActivityTurnShareRegist:
		a.ActivityTurnShareRegist(ctx)
	case *pb.ActivityTurnReq:
		a.ActivityTurnReq(ctx)
	case *pb.ActivityTurnClickLogReq:
		a.ActivityTurnClickLogReq(ctx)
	case *pb.ActivityTurnDrawReq:
		a.ActivityTurnDrawReq(ctx)
	case *pb.ActivityTurnDrawLogReq:
		a.ActivityTurnDrawLogReq(ctx)
	case *pb.ActivityTurnTackPrizeReq:
		a.ActivityTurnTackPrizeReq(ctx)
	case *pb.ActivityTurnPrizesReq:
		a.ActivityTurnPrizesReq(ctx)
	case *pb.ActivityTurnPrizeAuditSuccess:
		a.ActivityTurnPrizeAuditSuccess(ctx)
	case *pb.PayUTRBackReword:
		a.PayUTRBackReword(ctx)
	case *pb.UploadUtrReq:
		a.UploadUtrReq(ctx)
	case *pb.UploadUtrRsp:
		a.UploadUtrRsp(ctx)
	case *pb.CouponList:
		a.CouponList(ctx)
	case *pb.IssueCoupon:
		a.IssueCoupon(ctx)
	case *pb.GetCoupon:
		a.GetCoupon(ctx)
	case *pb.SeeCouponLog:
		a.SeeCouponLog(ctx)
	case *pb.ActivityPopNtf:
		a.ActivityPopNtf(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
