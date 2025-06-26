package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// Handler 消息处理
func (rs *RoleActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.OfflineStop:
		rs.OfflineStop(ctx)
	case *pb.ServeClose:
		rs.ServeClose(ctx)
	case *pb.ServeStop:
		rs.ServeStop(ctx)
	case *pb.ServeStoped:
	case *pb.ServeStart:
		rs.ServeStart(ctx)
	case *pb.ServeStarted:
	case *pb.Tick:
		rs.Tick(ctx)
	case *pb.CloseServer:
		rs.CloseServer(ctx)
	case *pb.PingReq:
		rs.PingReq(ctx)
	case proto.Message:
		//响应消息
		//rs.Send(msg)
		rs.handlerLogin(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家数据请求处理
func (rs *RoleActor) handlerLogin(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.RegistReq:
		rs.RegistReq(ctx)
	case *pb.LoginReq:
		rs.LoginReq(ctx)
	// case *pb.WxLoginReq:
	// 	rs.WxLoginReq(ctx)
	case *pb.ResetPwdReq:
		rs.ResetPwdReq(ctx)
	case *pb.TouristReq:
		rs.TouristReq(ctx)
	case *pb.LoginElse:
		rs.LoginElse(ctx)
	case *pb.LoginSuccess:
		rs.LoginSuccess(ctx)
	case *pb.AddBlackList:
		rs.AddBlackList(ctx)
	case *pb.AddWhiteList:
		rs.AddWhiteList(ctx)
	case proto.Message:
		if rs.User == nil {
			glog.Errorf("user empty message %v", msg)
			return
		}
		rs.handlerUser(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家数据请求处理
func (rs *RoleActor) handlerUser(ctx actor.Context) {
	switch msg := ctx.Message().(type) {

	case *pb.GetGameListReq:
		rs.GetGameListReq(ctx)
	// case *pb.NoticeReq:
	// 	rs.NoticeReq(ctx)
	// case *pb.NoticeRsp:
	// 	rs.NoticeRsp(ctx)
	// case *pb.ActivityReq:
	// 	rs.ActivityReq(ctx)
	// case *pb.ActivityRsp:
	// 	rs.ActivityRsp(ctx)
	// case *pb.JoinActivityReq:
	// 	rs.JoinActivityReq(ctx)
	case *pb.GetCurrencyReq:
		rs.GetCurrencyReq(ctx)
	// case *pb.BuyReq:
	// 	rs.BuyReq(ctx)
	case *pb.ShopReq:
		rs.ShopReq(ctx)
	case *pb.PublishPayChannelUpdate:
		rs.PublishPayChannelUpdate(ctx)
	// case *pb.BankReq:
	// 	rs.BankReq(ctx)
	// case *pb.RankReq:
	// 	rs.RankReq(ctx)
	// case *pb.BankLogReq:
	// 	rs.BankLogReq(ctx)
	// case *pb.TaskReq:
	// 	rs.TaskReq(ctx)
	// case *pb.LuckyReq:
	// 	rs.LuckyReq(ctx)
	// case *pb.TaskPrizeReq:
	// 	rs.TaskPrizeReq(ctx)
	// case *pb.LoginPrizeReq:
	// 	rs.LoginPrizeReq(ctx)
	case *pb.SignatureReq:
		rs.SignatureReq(ctx)
	case *pb.PhotoReq:
		rs.PhotoReq(ctx)
	case *pb.NickNameReq:
		rs.NickNameReq(ctx)
	// case *pb.LatLngReq:
	// 	rs.LatLngReq(ctx)
	// case *pb.RoomRecordReq:
	// 	rs.RoomRecordReq(ctx)
	case *pb.UserDataReq:
		rs.UserDataReq(ctx)
	case *pb.GotUserData:
	// 	rs.GotUserData(ctx)
	// case *pb.BroadcastNtf:
	// rs.BroadcastNtf(ctx)
	// case *pb.FirstRechargeReq:
	// 	rs.FirstRechargeReq(ctx)
	// case *pb.FirstRechargeRewardReq:
	// 	rs.FirstRechargeRewardReq(ctx)
	case *pb.RechargeActivityReq:
		rs.RechargeActivityReq(ctx)
	case *pb.WeeklyCardReq:
		rs.WeeklyCardReq(ctx)
	case *pb.DailySignDataReq:
		rs.DailySignDataReq(ctx)
	case *pb.DailySignReceiveReq:
		rs.DailySignReceiveReq(ctx)
	case *pb.BindPhoneReq:
		rs.BindPhoneReq(ctx)
	case *pb.WeeklyCardReceiveReq:
		rs.WeeklyCardReceiveReq(ctx)
	case *pb.OnlineRewardDataReq:
		rs.OnlineRewardDataReq(ctx)
	case *pb.OnlineRewardTimeReq:
		rs.OnlineRewardTimeReq(ctx)
	case *pb.OnlineRewardReceiveReq:
		rs.OnlineRewardReceiveReq(ctx)
	case *pb.GetWithDrawDataReq:
		rs.GetWithDrawDataReq(ctx)
	case *pb.WithdrawLogReq:
		rs.WithdrawLogReq(ctx)
	case *pb.FeedBackReq:
		rs.FeedBackReq(ctx)
	case *pb.ReadFeedBackLogReq:
		rs.ReadFeedBackLogReq(ctx)
	case *pb.CustomerSendReq:
		rs.CustomerSendReq(ctx)
	case *pb.CustomerUploadFile:
		rs.CustomerUploadFile(ctx)
	case *pb.ConsumerReplyMessage:
		rs.ConsumerReplyMessage(ctx)
	case *pb.PublishCustomerSessionOver:
		rs.PublishCustomerSessionOver(ctx)
	case *pb.PublishCustomerMessageRevoke:
		rs.PublishCustomerMessageRevoke(ctx)
	case *pb.CustomerReadAck2Req:
		rs.CustomerReadAck2Req(ctx)
	case *pb.CustomerHistoryReq:
		rs.CustomerHistoryReq(ctx)
	case *pb.CustomerScoreReq:
		rs.CustomerScoreReq(ctx)
	case *pb.RegistRewardReq:
		rs.RegistRewardReq(ctx)
	case *pb.BindLoginPhoneReq:
		rs.BindLoginPhoneReq(ctx)
	case *pb.CloseWithdrawReq:
		rs.CloseWithdrawReq(ctx)
	case *pb.ShareDataReq:
		rs.ShareDataReq(ctx)
	case *pb.ShareWithdrawReq:
		rs.ShareWithdrawReq(ctx)
	case *pb.ShareSettlement:
		rs.ShareSettlement(ctx)
	case *pb.ShareReferralReq:
		rs.ShareReferralReq(ctx)
	case *pb.ShareRewardReq:
		rs.ShareRewardReq(ctx)
	case *pb.ShareRewardDetailReq:
		rs.ShareRewardDetailReq(ctx)
	case *pb.GameGuidReq:
		rs.GameGuidReq(ctx)
	case *pb.BugFeedbackReq:
		rs.BugFeedbackReq(ctx)
	case *pb.ButtonClickReq:
		rs.ButtonClickReq(ctx)
	case *pb.ButtonClick2Req:
		rs.ButtonClick2Req(ctx)
	case *pb.RechargeReportReq:
		rs.RechargeReportReq(ctx)
	case *pb.SaveBankInfoReq:
		rs.SaveBankInfoReq(ctx)
	case *pb.AdjustAdidReq:
		rs.AdjustAdidReq(ctx)
	case *pb.AnnouncementReq:
		rs.AnnouncementReq(ctx)
	case *pb.CDKReq:
		rs.CDKReq(ctx)
	case *pb.GetUserInfoReq:
		rs.GetUserInfoReq(ctx)
	case *pb.ExternalBetReq:
		rs.ExternalBetReq(ctx)
	case *pb.ExternalRewardReq:
		rs.ExternalRewardReq(ctx)
	case *pb.EfiTransactionReq:
		rs.EfiTransactionReq(ctx)
	case *pb.BreakingGiftReq:
		rs.BreakingGiftReq(ctx)
	case *pb.ShopPotRewardReq:
		rs.ShopPotRewardReq(ctx)
	case *pb.VipRewardReq:
		rs.VipRewardReq(ctx)
	case *pb.RankWithdrawReq:
		rs.RankWithdrawReq(ctx)
	case *pb.LuckyDrawReq:
		rs.LuckyDrawReq(ctx)
	case *pb.LuckyDrawHistoryReq:
		rs.LuckyDrawHistoryReq(ctx)
	case *pb.LuckyDrawTakeReq:
		rs.LuckyDrawTakeReq(ctx)
	case *pb.LuckyDrawNumberMail:
		rs.LuckyDrawNumberMail(ctx)
	case *pb.ScratchTicketRewardReq:
		rs.ScratchTicketRewardReq(ctx)
	case *pb.ActivityMonitorReq:
		rs.ActivityMonitorReq(ctx)
	case *pb.PlayShareDrawReq:
		rs.PlayShareDrawReq(ctx)
	case *pb.PlayShareRewardReq:
		rs.PlayShareRewardReq(ctx)
	case *pb.PlayshareTest:
		rs.PlayshareTest()
	case *pb.VBTaskRewardReq:
		rs.VBTaskRewardReq(ctx)
	case *pb.GetVBTaskReq:
		rs.GetVBTaskReq(ctx)
	case *pb.VBLoseCompensationReq:
		rs.VBLoseCompensationReq(ctx)
	case *pb.VBLoseCompensationRewardReq:
		rs.VBLoseCompensationRewardReq(ctx)
	case *pb.WithdrawalPopReq:
		rs.WithdrawalPopReq(ctx)
	case *pb.VipBonusReq:
		rs.VipBonusReq(ctx)
	case *pb.VipPopReq:
		rs.VipPopReq(ctx)
	case *pb.VipBankLogRecordReq:
		rs.VipBankLogRecordReq(ctx)
	case *pb.ActivityBetRankReq:
		rs.ActivityBetRankReq(ctx)
	case *pb.ActivityBetRankHistoryReq:
		rs.ActivityBetRankHistoryReq(ctx)
	case *pb.ActivityBetRankMyRecordReq:
		rs.ActivityBetRankMyRecordReq(ctx)
	case *pb.ActivityBetRankNewRankSkipTodayReq:
		rs.ActivityBetRankNewRankSkipTodayReq(ctx)
	case *pb.ActivityBetRankWillRankSkipTodayReq:
		rs.ActivityBetRankWillRankSkipTodayReq(ctx)
	case *pb.ActivityBetRankLoseRankSkipTodayReq:
		rs.ActivityBetRankLoseRankSkipTodayReq(ctx)
	case *pb.ActivityBetRankRewordReq:
		rs.ActivityBetRankRewordReq(ctx)
	case *pb.ActivityBetRankRewordNtf:
		rs.Send(msg)
	case *pb.ActivityBetRankNewRankNtf:
		rs.Send(msg)
	case *pb.ActivityBetRankWillRankNtf:
		rs.Send(msg)
	case *pb.ActivityBetRankLoseRankNtf:
		rs.Send(msg)
	case *pb.ActivityBetRankUpdateNtf:
		rs.Send(msg)
	case *pb.UpdateFavoriteReq:
		rs.UpdateFavoriteReq(ctx)
	case *pb.ActivityTurnSync:
		rs.ActivityTurnSync(ctx)
	case *pb.ActivityTurnReq:
		rs.ActivityTurnReq(ctx)
	case *pb.ActivityTurnClickLogReq:
		rs.ActivityTurnClickLogReq(ctx)
	case *pb.ActivityTurnDrawReq:
		rs.ActivityTurnDrawReq(ctx)
	case *pb.ActivityTurnDrawLogReq:
		rs.ActivityTurnDrawLogReq(ctx)
	case *pb.ActivityTurnTackPrizeReq:
		rs.ActivityTurnTackPrizeReq(ctx)
	case *pb.ActivityTurnPrizesReq:
		rs.ActivityTurnPrizesReq(ctx)
	case *pb.BalanceRecordsReq:
		rs.BalanceRecordsReq(ctx)
	case *pb.SkipGiftWelfareReq:
		rs.SkipGiftWelfareReq(ctx)
	case *pb.LaunchInviteLogReq:
		rs.LaunchInviteLogReq(ctx)
	// ============================= 代理活动start
	case *pb.ShareAgentSync:
		rs.ShareAgentSync(ctx)
	case *pb.ShareOngoingEarningReq:
		rs.ShareOngoingEarningReq(ctx)
	case *pb.ShareInstantBonusReq:
		rs.ShareInstantBonusReq(ctx)
	case *pb.ShareAgentIncomeRecordBettingCommissionReq:
		rs.ShareAgentIncomeRecordBettingCommissionReq(ctx)
	case *pb.ShareAgentIncomeRecordReferralBonusReq:
		rs.ShareAgentIncomeRecordReferralBonusReq(ctx)
	case *pb.ShareAgentIncomeRecordAgentDetailsReq:
		rs.ShareAgentIncomeRecordAgentDetailsReq(ctx)
	case *pb.ShareAgentRulesReq:
		rs.ShareAgentRulesReq(ctx)
	case *pb.ShareOngoingEarningUnclaimedTackReq:
		rs.ShareOngoingEarningUnclaimedTackReq(ctx)
	case *pb.ShareInstantEarningUnclaimedTackReq:
		rs.ShareInstantEarningUnclaimedTackReq(ctx)
	case *pb.ShareAgentIncomeTackReq:
		rs.ShareAgentIncomeTackReq(ctx)
	// ============================= 代理活动end
	case *pb.GiftPackCodeTackReq:
		rs.GiftPackCodeTackReq(ctx)
	case *pb.VolatilitySubsidyWheelReq:
		rs.VolatilitySubsidyWheelReq(ctx)
	case *pb.VolatilitySubsidySync:
		rs.VolatilitySubsidySync(ctx)
	case proto.Message:
		rs.handlerInternal(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 内部转发
func (rs *RoleActor) handlerInternal(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.ShareUserRegist:
		rs.ShareUserRegist(ctx)
	case *pb.FeedBackReply:
		rs.FeedBackReply(ctx)
	case *pb.ChangeUser:
		rs.ChangeUser(ctx)
	case *pb.PointControl:
		rs.PointControl(ctx)
	case *pb.ChangePointControl:
		rs.ChangePointControl(ctx)
	case *pb.OnlineUser:
		rs.OnlineUser(ctx)
	case *pb.EventPost:
		rs.EventPost(ctx)
	case *pb.LuckyUpdate:
		rs.LuckyUpdate(ctx)
	case *pb.BankGive:
		rs.BankGive(ctx)
	case *pb.TaskUpdate:
		rs.TaskUpdate(ctx)
	case *pb.GiveAndOutCash:
		rs.GiveAndOutCash(ctx)
	case *pb.ChangeShareSuperior:
		rs.ChangeShareSuperior(ctx)
	case *pb.RemoveShareJuntor:
		rs.RemoveShareJuntor(ctx)
	case *pb.ChangePhone:
		rs.ChangePhone(ctx)
	case *pb.ChangeViceAccount:
		rs.ChangeViceAccount(ctx)
	case *pb.ClearBankInfo:
		rs.ClearBankInfo(ctx)
	case *pb.GiveWithdrawCash:
		rs.GiveWithdrawCash(ctx)
	case *pb.MarQueeWithdraw:
		rs.MarQueeWithdraw(ctx)
	case *pb.UserCustomPhoto:
		rs.UserCustomPhoto(ctx)
	case proto.Message:
		rs.handlerTask(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 任务数据请求
func (rs *RoleActor) handlerTask(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	// case *pb.PokerHandsTaskReq:
	// 	rs.PokerHandsTaskReq(ctx)
	// case *pb.PokerHandsTaskGetReq:
	// 	rs.PokerHandsRewardReq(ctx)
	case *pb.GetTaskReq:
		rs.GetTaskReq(ctx)
	case *pb.GetTaskRewardReq:
		rs.GetTaskRewardReq(ctx)
	case proto.Message:
		rs.handlerPay(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 充值数据请求处理
func (rs *RoleActor) handlerPay(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	// case *pb.ApplePayReq:
	// 	rs.ApplePayReq(ctx)
	// case *pb.WxpayOrderReq:
	// 	rs.WxpayOrderReq(ctx)
	// case *pb.WxpayQueryReq:
	// 	rs.WxpayQueryReq(ctx)
	case *pb.PayGoods:
		rs.PayGoods(ctx)
	case *pb.PayCurrency:
		rs.PayCurrency(ctx)
	case *pb.ChangeCurrency:
		rs.ChangeCurrency(ctx)
	case *pb.ModifyCurrency:
		rs.ModifyCurrency(ctx)
	case *pb.ExternalModifyCurrencyReq:
		rs.ExternalModifyCurrencyReq(ctx)
	case *pb.PayOrderReq:
		rs.PayOrderReq(ctx)
	case *pb.PayWithDrawReq:
		rs.WithDrawReq(ctx)
	case *pb.PakWithDrawReq:
		rs.PakWithDrawReq(ctx)
	case *pb.WithdrawLogStatus:
		rs.WithdrawLogStatus(ctx)
	case *pb.ShareRecharge:
		rs.ShareRecharge(ctx)
	case *pb.PaymentRecordReq:
		rs.PaymentRecordReq(ctx)
	case *pb.PaymentRecordUTRReq:
		rs.PaymentRecordUTRReq(ctx)
	case *pb.WithdrawRecordReq:
		rs.WithdrawRecordReq(ctx)
	case *pb.WithdrawRecordRefundReq:
		rs.WithdrawRecordRefundReq(ctx)
	case *pb.WithdrawRecordKeepWaitingReq:
		rs.WithdrawRecordKeepWaitingReq(ctx)
	case *pb.WithdrawRecordOvertimePayTackReq:
		rs.WithdrawRecordOvertimePayTackReq(ctx)
	case *pb.WithdrawBankInfoReq:
		rs.WithdrawBankInfoReq(ctx)
	case *pb.CouponReq:
		rs.CouponReq()
	case *pb.IssueCouponNtf:
		rs.IssueCouponNtf(ctx)
	case *pb.SeeCouponReq:
		rs.SeeCouponReq(ctx)
	case proto.Message:
		rs.handlerDesk(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 玩家桌子常用共有操作请求处理
func (rs *RoleActor) handlerDesk(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.EnteredDesk:
		rs.EnteredDesk(ctx)
	case *pb.MatchedDesk:
		rs.MatchedDesk(ctx)
	case *pb.CreatedDesk:
		rs.CreatedDesk(ctx)
	case *pb.LeftDesk:
		rs.LeftDesk(ctx)
	case *pb.SetRecord:
		rs.SetRecord(ctx)
	case *pb.GotRoomList:
		rs.GotRoomList(ctx)
	case *pb.ChangedDesk:
		rs.ChangedDesk(ctx)
	case *pb.FreeSetRecord:
		rs.FreeSetRecord(ctx)
	case *pb.LogGameTime:
		rs.LogGameTime(ctx)
	case *pb.TPTotalRound:
		rs.TPTotalRound(ctx)
	case *pb.TPTriggeTimes:
		rs.TPTriggeTimes(ctx)
	case *pb.TriggerStategy100:
		rs.TriggerStategy100(ctx)
	case *pb.TriggerStategy200:
		rs.TriggerStategy200(ctx)
	case *pb.ChatTextReq:
		rs.ChatTextReq(ctx)
	case *pb.HedgeTrigger:
		rs.HedgeTrigger(ctx)
	case *pb.KickoutedRoom:
		rs.KickoutedRoom(ctx)
	case *pb.TriggerFreeWelfare:
		rs.TriggerFreeWelfare(ctx)
	case *pb.TpUserStageChangeSync:
		rs.TpUserStageChangeSync(ctx)
	case *pb.TpUserRoundSync:
		rs.TpUserRoundSync(ctx)
	case *pb.TpUserModelSync:
		rs.TpUserModelSync(ctx)
	case *pb.TpWinOrLoseAmountSync:
		rs.TpWinOrLoseAmountSync(ctx)
	case *pb.TpUserChangeCorrectionValueSync:
		rs.TpUserChangeCorrectionValueSync(ctx)
	case *pb.TpUserTotalWinOrLoseAmountSync:
		rs.TpUserTotalWinOrLoseAmountSync(ctx)
	case *pb.TpUserChargeInGameNumOfTriggerSync:
		rs.TpUserChargeInGameNumOfTriggerSync(ctx)
	case *pb.TpUserChargeInGameNumOfSuccessSync:
		rs.TpUserChargeInGameNumOfSuccessSync(ctx)
	case *pb.TpUserChargeInGameStorySync:
		rs.TpUserChargeInGameStorySync(ctx)
	case *pb.TpUserStoryPlusSync:
		rs.TpUserStoryPlusSync(ctx)
	case *pb.TpUserFollowRateTriggerSync:
		rs.TpUserFollowRateTriggerSync(ctx)
	case *pb.TpUserFollowRateSuccessSync:
		rs.TpUserFollowRateSuccessSync(ctx)
	case *pb.TpUserStoryCDSync:
		rs.TpUserStoryCDSync(ctx)
	case *pb.TpUserStoryCDReduce:
		rs.TpUserStoryCDReduce(ctx)
	case *pb.TpUserControlStrategyHistorySync:
		rs.TpUserControlStrategyHistorySync(ctx)
	case *pb.TpUserGameTimeSync:
		rs.TpUserGameTimeSync(ctx)
	case *pb.TpUserTodayGameRoundSync:
		rs.TpUserTodayGameRoundSync(ctx)
	case *pb.TpUserTodayControlStrategyNumSync:
		rs.TpUserTodayControlStrategyNumSync(ctx)
	case proto.Message:
		rs.handlerHua(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// hua请求处理
func (rs *RoleActor) handlerHua(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.JHCoinEnterRoomReq:
		rs.JHCoinEnterRoomReq(ctx)
	case *pb.JHFreeEnterRoomReq:
		rs.JHFreeEnterRoomReq(ctx)
	case *pb.JHFreeDealerReq:
		rs.JHFreeDealerReq(ctx)
	case *pb.JHFreeDealerListReq:
		rs.JHFreeDealerListReq(ctx)
	case *pb.JHSitReq:
		rs.JHSitReq(ctx)
	case *pb.JHFreeBetReq:
		rs.JHFreeBetReq(ctx)
	case *pb.JHFreeTrendReq:
		rs.JHFreeTrendReq(ctx)
	case *pb.JHFreeWinersReq:
		rs.JHFreeWinersReq(ctx)
	case *pb.JHFreeRolesReq:
		rs.JHFreeRolesReq(ctx)
	case *pb.JHRoomListReq:
		rs.JHRoomListReq(ctx)
	case *pb.JHEnterRoomReq:
		rs.JHEnterRoomReq(ctx)
	case *pb.JHCreateRoomReq:
		rs.JHCreateRoomReq(ctx)
	case *pb.JHLeaveReq:
		rs.JHLeaveReq(ctx)
	case *pb.JHReadyReq:
		rs.JHReadyReq(ctx)
	case *pb.JHReady2Req:
		rs.JHReady2Req(ctx)
	case *pb.JHGameStartReq:
		rs.JHGameStartReq(ctx)
	// case *pb.ChatTextReq:
	// 	rs.ChatTextReq(ctx)
	case *pb.JHGameRecordReq:
		rs.JHGameRecordReq(ctx)
	case *pb.JHLaunchVoteReq:
		rs.JHLaunchVoteReq(ctx)
	case *pb.JHVoteReq:
		rs.JHVoteReq(ctx)
	case *pb.JHCoinSeeReq:
		rs.JHCoinSeeReq(ctx)
	case *pb.JHCoinCallReq:
		rs.JHCoinCallReq(ctx)
	case *pb.JHCoinRaiseReq:
		rs.JHCoinRaiseReq(ctx)
	case *pb.JHCoinFoldReq:
		rs.JHCoinFoldReq(ctx)
	case *pb.JHCoinBiReq:
		rs.JHCoinBiReq(ctx)
	case *pb.JHCoinReplyBiReq:
		rs.JHCoinReplyBiReq(ctx)
	case *pb.JHCoinChangeRoomReq:
		rs.JHCoinChangeRoomReq(ctx)
	case *pb.JHKickoutRoomReq:
		rs.JHKickoutRoomReq(ctx)
	case *pb.JHChangeSeatReq:
		rs.JHChangeSeatReq(ctx)
	case *pb.JHSetTpNewbieProbe:
		rs.JHSetTpNewbieProbe(ctx)
	case *pb.JHSetTpHurtScore:
		rs.JHSetTpHurtScore(ctx)
	case *pb.JHVHRecordSync:
		rs.JHVHRecordSync(ctx)
	case *pb.JHHeartbeatSync:
		rs.JHHeartbeatSync(ctx)
	case *pb.JHUserLjsbPyCDSync:
		rs.JHUserLjsbPyCDSync(ctx)
	case *pb.JHUserGCYXSync:
		rs.JHUserGCYXSync(ctx)
	case proto.Message:
		rs.handlerJOKER(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// joker请求处理
func (rs *RoleActor) handlerJOKER(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.JOKERCoinEnterRoomReq:
		rs.JOKERCoinEnterRoomReq(ctx)
	case *pb.JOKERFreeEnterRoomReq:
		rs.JOKERFreeEnterRoomReq(ctx)
	case *pb.JOKERFreeDealerReq:
		rs.JOKERFreeDealerReq(ctx)
	case *pb.JOKERFreeDealerListReq:
		rs.JOKERFreeDealerListReq(ctx)
	case *pb.JOKERSitReq:
		rs.JOKERSitReq(ctx)
	case *pb.JOKERFreeBetReq:
		rs.JOKERFreeBetReq(ctx)
	case *pb.JOKERFreeTrendReq:
		rs.JOKERFreeTrendReq(ctx)
	case *pb.JOKERFreeWinersReq:
		rs.JOKERFreeWinersReq(ctx)
	case *pb.JOKERFreeRolesReq:
		rs.JOKERFreeRolesReq(ctx)
	case *pb.JOKERRoomListReq:
		rs.JOKERRoomListReq(ctx)
	case *pb.JOKEREnterRoomReq:
		rs.JOKEREnterRoomReq(ctx)
	case *pb.JOKERCreateRoomReq:
		rs.JOKERCreateRoomReq(ctx)
	case *pb.JOKERLeaveReq:
		rs.JOKERLeaveReq(ctx)
	case *pb.JOKERReadyReq:
		rs.JOKERReadyReq(ctx)
	case *pb.JOKERReady2Req:
		rs.JOKERReady2Req(ctx)
	// case *pb.ChatTextReq:
	// 	rs.ChatTextReq(ctx)
	case *pb.JOKERGameRecordReq:
		rs.JOKERGameRecordReq(ctx)
	case *pb.JOKERLaunchVoteReq:
		rs.JOKERLaunchVoteReq(ctx)
	case *pb.JOKERVoteReq:
		rs.JOKERVoteReq(ctx)
	case *pb.JOKERCoinSeeReq:
		rs.JOKERCoinSeeReq(ctx)
	case *pb.JOKERCoinCallReq:
		rs.JOKERCoinCallReq(ctx)
	case *pb.JOKERCoinRaiseReq:
		rs.JOKERCoinRaiseReq(ctx)
	case *pb.JOKERCoinFoldReq:
		rs.JOKERCoinFoldReq(ctx)
	case *pb.JOKERCoinBiReq:
		rs.JOKERCoinBiReq(ctx)
	case *pb.JOKERCoinReplyBiReq:
		rs.JOKERCoinReplyBiReq(ctx)
	case *pb.JOKERCoinChangeRoomReq:
		rs.JOKERCoinChangeRoomReq(ctx)
	case proto.Message:
		rs.handlerAK47(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// ak47请求处理
func (rs *RoleActor) handlerAK47(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.AK47CoinEnterRoomReq:
		rs.AK47CoinEnterRoomReq(ctx)
	case *pb.AK47FreeEnterRoomReq:
		rs.AK47FreeEnterRoomReq(ctx)
	case *pb.AK47FreeDealerReq:
		rs.AK47FreeDealerReq(ctx)
	case *pb.AK47FreeDealerListReq:
		rs.AK47FreeDealerListReq(ctx)
	case *pb.AK47SitReq:
		rs.AK47SitReq(ctx)
	case *pb.AK47FreeBetReq:
		rs.AK47FreeBetReq(ctx)
	case *pb.AK47FreeTrendReq:
		rs.AK47FreeTrendReq(ctx)
	case *pb.AK47FreeWinersReq:
		rs.AK47FreeWinersReq(ctx)
	case *pb.AK47FreeRolesReq:
		rs.AK47FreeRolesReq(ctx)
	case *pb.AK47RoomListReq:
		rs.AK47RoomListReq(ctx)
	case *pb.AK47EnterRoomReq:
		rs.AK47EnterRoomReq(ctx)
	case *pb.AK47CreateRoomReq:
		rs.AK47CreateRoomReq(ctx)
	case *pb.AK47LeaveReq:
		rs.AK47LeaveReq(ctx)
	case *pb.AK47ReadyReq:
		rs.AK47ReadyReq(ctx)
	case *pb.AK47Ready2Req:
		rs.AK47Ready2Req(ctx)
	// case *pb.ChatTextReq:
	// 	rs.ChatTextReq(ctx)
	case *pb.AK47GameRecordReq:
		rs.AK47GameRecordReq(ctx)
	case *pb.AK47LaunchVoteReq:
		rs.AK47LaunchVoteReq(ctx)
	case *pb.AK47VoteReq:
		rs.AK47VoteReq(ctx)
	case *pb.AK47CoinSeeReq:
		rs.AK47CoinSeeReq(ctx)
	case *pb.AK47CoinCallReq:
		rs.AK47CoinCallReq(ctx)
	case *pb.AK47CoinRaiseReq:
		rs.AK47CoinRaiseReq(ctx)
	case *pb.AK47CoinFoldReq:
		rs.AK47CoinFoldReq(ctx)
	case *pb.AK47CoinBiReq:
		rs.AK47CoinBiReq(ctx)
	case *pb.AK47CoinReplyBiReq:
		rs.AK47CoinReplyBiReq(ctx)
	case *pb.AK47CoinChangeRoomReq:
		rs.AK47CoinChangeRoomReq(ctx)
	case proto.Message:
		rs.handlerRM(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// rm请求处理
func (rs *RoleActor) handlerRM(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	// case *pb.RMRoomListReq:
	// 	rs.RMRoomListReq(ctx)
	case *pb.RMCoinEnterRoomReq:
		rs.RMCoinEnterRoomReq(ctx)
	case *pb.RMLeaveReq:
		rs.RMLeaveReq(ctx)
	case *pb.RMReady2Req:
		rs.RMReady2Req(ctx)
	case *pb.RMGameStartReq:
		rs.RMGameStartReq(ctx)
	case *pb.RMDrawCardReq:
		rs.RMDrawCardReq(ctx)
	case *pb.RMDiscardReq:
		rs.RMDiscardReq(ctx)
	case *pb.RMFinishReq:
		rs.RMFinishReq(ctx)
	case *pb.RMDeclareReq:
		rs.RMDeclareReq(ctx)
	case *pb.RMSortReq:
		rs.RMSortReq(ctx)
	case *pb.RMDropReq:
		rs.RMDropReq(ctx)
	case *pb.RMDropScoreReq:
		rs.RMDropScoreReq(ctx)
	// case *pb.ChatTextReq:
	// 	rs.ChatTextReq(ctx)
	// case *pb.RMFreeEnterRoomReq:
	// 	rs.RMFreeEnterRoomReq(ctx)
	// case *pb.RMFreeDealerReq:
	// 	rs.RMFreeDealerReq(ctx)
	// case *pb.RMFreeDealerListReq:
	// 	rs.RMFreeDealerListReq(ctx)
	// case *pb.RMSitReq:
	// 	rs.RMSitReq(ctx)
	// case *pb.RMFreeBetReq:
	// 	rs.RMFreeBetReq(ctx)
	// case *pb.RMFreeTrendReq:
	// 	rs.RMFreeTrendReq(ctx)
	// case *pb.RMFreeWinersReq:
	// 	rs.RMFreeWinersReq(ctx)
	// case *pb.RMFreeRolesReq:
	// 	rs.RMFreeRolesReq(ctx)
	case *pb.RMEnterRoomReq:
		rs.RMEnterRoomReq(ctx)
	case *pb.RMCreateRoomReq:
		rs.RMCreateRoomReq(ctx)
	case *pb.RMKickoutRoomReq:
		rs.RMKickoutRoomReq(ctx)
	case *pb.RMChangeSeatReq:
		rs.RMChangeSeatReq(ctx)
	// case *pb.RMReadyReq:
	// 	rs.RMReadyReq(ctx)

	// case *pb.ChatTextReq:
	// 	rs.ChatTextReq(ctx)
	// case *pb.RMGameRecordReq:
	// 	rs.RMGameRecordReq(ctx)
	case *pb.RMLaunchVoteReq:
		rs.RMLaunchVoteReq(ctx)
	case *pb.RMVoteReq:
		rs.RMVoteReq(ctx)
	case *pb.RMAutoSortReq:
		rs.RMAutoSortReq(ctx)
	case *pb.RMQiCardsReq:
		rs.RMQiCardsReq(ctx)
	case *pb.RMRoiRecordsync:
		rs.RMRoiRecordsync(ctx)
	// case *pb.RMCoinSeeReq:
	// 	rs.RMCoinSeeReq(ctx)
	// case *pb.RMCoinCallReq:
	// 	rs.RMCoinCallReq(ctx)
	// case *pb.RMCoinRaiseReq:
	// 	rs.RMCoinRaiseReq(ctx)
	// case *pb.RMCoinFoldReq:
	// 	rs.RMCoinFoldReq(ctx)
	// case *pb.RMCoinBiReq:
	// 	rs.RMCoinBiReq(ctx)
	// case *pb.RMCoinReplyBiReq:
	// 	rs.RMCoinReplyBiReq(ctx)
	// case *pb.RMCoinChangeRoomReq:
	// 	rs.RMCoinChangeRoomReq(ctx)
	case proto.Message:
		rs.handlerLhd(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// lhd请求处理
func (rs *RoleActor) handlerLhd(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.LHFreeEnterRoomReq:
		rs.LHFreeEnterRoomReq(ctx)
	case *pb.LHFreeBetReq:
		rs.LHFreeBetReq(ctx)
	case *pb.LHRoomListReq:
		rs.LHRoomListReq(ctx)
	case *pb.LHLeaveReq:
		rs.LHLeaveReq(ctx)
	case *pb.LHEnterSuccessReq:
		rs.LHEnterSuccessReq(ctx)
	case *pb.LHGCYXSync:
		rs.LHGCYXSync(ctx)
	case proto.Message:
		rs.handler7up(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 7up请求处理
func (rs *RoleActor) handler7up(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.UPFreeEnterRoomReq:
		rs.UPFreeEnterRoomReq(ctx)
	case *pb.UPEnterSuccessReq:
		rs.UPEnterSuccessReq(ctx)
	case *pb.UPDoubleBetReq:
		rs.UPDoubleBetReq(ctx)
	case *pb.UPUndoBetReq:
		rs.UPUndoBetReq(ctx)
	case *pb.UPFreeBetReq:
		rs.UPFreeBetReq(ctx)
	case *pb.UPRoomListReq:
		rs.UPRoomListReq(ctx)
	case *pb.UPLeaveReq:
		rs.UPLeaveReq(ctx)
	case *pb.UPGCYXSync:
		rs.UPGCYXSync(ctx)
	case *pb.UPMyHistoryReq:
		rs.UPMyHistoryReq(ctx)
	case proto.Message:
		rs.handlerCrash(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// crash请求处理
func (rs *RoleActor) handlerCrash(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.CRASHEnterRoomReq:
		rs.CRASHEnterRoomReq(ctx)
	case *pb.CRASHEnterSuccessReq:
		rs.CRASHEnterSuccessReq(ctx)
	case *pb.CRASHBetReq:
		rs.CRASHBetReq(ctx)
	case *pb.CRASHCancelBetReq:
		rs.CRASHCancelBetReq(ctx)
	case *pb.CRASHBackReq:
		rs.CRASHBackReq(ctx)
	case *pb.CRASHRoomListReq:
		rs.CRASHRoomListReq(ctx)
	case *pb.CRASHLeaveReq:
		rs.CRASHLeaveReq(ctx)
	case *pb.CRASHCrashMultipleReq:
		rs.CRASHCrashMultipleReq(ctx)
	case *pb.CrashMyHistoryReq:
		rs.CrashMyHistoryReq(ctx)
	case *pb.CrashTopWinnersReq:
		rs.CrashTopWinnersReq(ctx)
	case *pb.CRASHEnterRoomRsp:
		rs.Send(msg)
	case *pb.CRASHEnterSuccessRsp:
		rs.Send(msg)
	case proto.Message:
		rs.handlerAB(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// ab请求处理
func (rs *RoleActor) handlerAB(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.ABFreeEnterRoomReq:
		rs.ABFreeEnterRoomReq(ctx)
	case *pb.ABFreeBetReq:
		rs.ABFreeBetReq(ctx)
	case *pb.ABRoomListReq:
		rs.ABRoomListReq(ctx)
	case *pb.ABLeaveReq:
		rs.ABLeaveReq(ctx)
	case *pb.ABCreateRoomReq:
		rs.ABCreateRoomReq(ctx)
	case *pb.ABEnterRoomReq:
		rs.ABEnterRoomReq(ctx)
	case *pb.ABKickoutRoomReq:
		rs.ABKickoutRoomReq(ctx)
	case *pb.ABChangeSeatReq:
		rs.ABChangeSeatReq(ctx)
	case *pb.ABChangeSeatAcceptReq:
		rs.ABChangeSeatAcceptReq(ctx)
	case *pb.ABLaunchVoteReq:
		rs.ABLaunchVoteReq(ctx)
	case *pb.ABVoteReq:
		rs.ABVoteReq(ctx)
	case *pb.ABBetReq:
		rs.ABBetReq(ctx)
	case *pb.ABGameStartReq:
		rs.ABGameStartReq(ctx)
	case proto.Message:
		rs.handlerLottery(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 彩票请求处理
func (rs *RoleActor) handlerLottery(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.LotteryEnterReq:
		rs.LotteryEnterReq(ctx)
	case *pb.LotteryBetReq:
		rs.LotteryBetReq(ctx)
	case *pb.LotteryLeaveReq:
		rs.LotteryLeaveReq(ctx)
	case proto.Message:
		rs.handlerPriv(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 私人房消息
func (rs *RoleActor) handlerPriv(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.PrivEnterRoomReq:
		rs.PrivEnterRoomReq(ctx)
	case *pb.PrivLaunchAgainReq:
		rs.PrivLaunchAgainReq(ctx)
	case *pb.PrivAgainReq:
		rs.PrivAgainReq(ctx)
	case proto.Message:
		rs.handlerPlane(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// plane请求处理
func (rs *RoleActor) handlerPlane(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.PLANEEnterRoomReq:
		rs.PLANEEnterRoomReq(ctx)
	case *pb.PLANEBetReq:
		rs.PLANEBetReq(ctx)
	case *pb.PLANECancelBetReq:
		rs.PLANECancelBetReq(ctx)
	case *pb.PLANEBackReq:
		rs.PLANEBackReq(ctx)
	case *pb.PLANERoomListReq:
		rs.PLANERoomListReq(ctx)
	case *pb.PLANELeaveReq:
		rs.PLANELeaveReq(ctx)
	case *pb.PLANECrashMultipleReq:
		rs.PLANECrashMultipleReq(ctx)
	case *pb.PLANELastRoundRecordReq:
		rs.PLANELastRoundRecordReq(ctx)
	case *pb.PLANEMyHistoryReq:
		rs.PLANEMyHistoryReq(ctx)
	case *pb.PLANETopWinnersReq:
		rs.PLANETopWinnersReq(ctx)
	case *pb.PLANEEnterRoomRsp:
		// msg.AutoCrash = rs.User.PlaneAutoLeave
		// msg.CrashMultiple = rs.User.PlaneMultiple
		rs.Send(msg)
	case proto.Message:
		rs.handlerRB(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// redblack请求处理
func (rs *RoleActor) handlerRB(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.RBFreeEnterRoomReq:
		rs.RBFreeEnterRoomReq(ctx)
	case *pb.RBFreeBetReq:
		rs.RBFreeBetReq(ctx)
	case *pb.RBLeaveReq:
		rs.RBLeaveReq(ctx)
	case proto.Message:
		rs.handlerMines(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// mines请求处理
func (rs *RoleActor) handlerMines(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.MinesEnterRoomReq:
		rs.MinesEnterRoomReq(ctx)
	case *pb.MinesBetReq:
		rs.MinesBetReq(ctx)
	case *pb.MinesPitStepReq:
		rs.MinesPitStepReq(ctx)
	case *pb.MinesCashOutReq:
		rs.MinesCashOutReq(ctx)
	case *pb.MinesAutoBetCancelReq:
		rs.MinesAutoBetCancelReq(ctx)
	case *pb.MinesLeaveReq:
		rs.MinesLeaveReq(ctx)
	case *pb.MinesMyHistoryReq:
		rs.MinesMyHistoryReq(ctx)
	case proto.Message:
		rs.handlerFortuneGems2(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// fortune_gems2请求处理
func (rs *RoleActor) handlerFortuneGems2(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.FortuneGems2EnterRoomReq:
		rs.FortuneGems2EnterRoomReq(ctx)
	case *pb.FortuneGems2BetReq:
		rs.FortuneGems2BetReq(ctx)
	case *pb.FortuneGems2LeaveReq:
		rs.FortuneGems2LeaveReq(ctx)
	// case *pb.FortuneGems2PitStepReq:
	// 	rs.FortuneGems2PitStepReq(ctx)
	// case *pb.FortuneGems2CashOutReq:
	// 	rs.FortuneGems2CashOutReq(ctx)
	// case *pb.FortuneGems2AutoBetCancelReq:
	// 	rs.FortuneGems2AutoBetCancelReq(ctx)
	// case *pb.FortuneGems2MyHistoryReq:
	// 	rs.FortuneGems2MyHistoryReq(ctx)
	case proto.Message:
		rs.handlerFortuneGems(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// fortune_gems请求处理
func (rs *RoleActor) handlerFortuneGems(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.FortuneGemsEnterRoomReq:
		rs.FortuneGemsEnterRoomReq(ctx)
	case *pb.FortuneGemsBetReq:
		rs.FortuneGemsBetReq(ctx)
	case *pb.FortuneGemsLeaveReq:
		rs.FortuneGemsLeaveReq(ctx)
	case proto.Message:
		//响应消息
		if rs.online {
			rs.Send(msg)
		}
	default:
		glog.Errorf("unknown message %v", msg)
	}
}
