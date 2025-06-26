package ck

import (
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/zlog"
	"runtime/debug"
)

var (
	natsTopics = []string{
		mq.TopicSyncUser,
		mq.TopicSyncUserFinance,
		mq.TopicSyncTradeRecord,
		mq.TopicSyncWithdrawRecord,
		mq.TopicSyncWithdrawRecordTransferOrderDetail,
		mq.TopicSyncLogWater,
		mq.TopicSyncDetail,
		mq.TopicSyncLogLogin,
		mq.TopicSyncExternalBet,
		mq.TopicSyncExternalReward,
		mq.TopicSyncExternalCancel,
		mq.TopicSyncActivityBetRankPrize,
		mq.TopicSyncActivityTurnDrawLog,
		mq.TopicSyncActivityTurnPrizeLog,
		mq.TopicSyncLaunchInviteLog,
		mq.TopicSyncShareAgentIncomeRecord,
		mq.TopicSyncShareAgentIncomeRecordTackLog,
		mq.TopicSyncLogVBDiamond,
		mq.TopicSyncLogOutDiamond,
		mq.TopicSyncVolatilitySubsidy,
		mq.TopicSyncOnlineUsersLog,
		mq.TopicSyncUserCoupon,
		mq.TopicSyncCustomerChatSession,
		mq.TopicSyncCustomerChatMessage,
		mq.TopicSyncCustomerSeatLog,
	}
)

type NatsDataProducer struct{}

func (NatsDataProducer) PublishDatas(topic string, datas []interface{}) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("nats PublishDatas error:", r)
			zlog.Error("nats PublishDatas error:", r)
			debug.PrintStack()
		}
	}()

	if len(datas) == 0 {
		return
	}

	// 遍历数据并发布
	for _, data := range datas {
		err := mq.NatsPublish(topic, data)
		if err != nil {
			glog.Errorf("nats publish error: topic=%s, error=%v", topic, err)
			zlog.Errorf("nats publish error: topic=%s, error=%v", topic, err)
			continue
		}
	}
}

func InitNatsProducer() {
	data.DataProducers = &NatsDataProducer{}
}

func InitNatsConsumer() {
	consumerName := "sync"

	for _, topic := range natsTopics {
		switch topic {
		case mq.TopicSyncUser:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.User { return new(data.User) },
				func(msg *data.User) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncUserFinance:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.UserFinance { return new(data.UserFinance) },
				func(msg *data.UserFinance) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncTradeRecord:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.TradeRecord { return new(data.TradeRecord) },
				func(msg *data.TradeRecord) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncWithdrawRecord:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.WithdrawRecord { return new(data.WithdrawRecord) },
				func(msg *data.WithdrawRecord) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncWithdrawRecordTransferOrderDetail:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.TransferOrderDetail { return new(data.TransferOrderDetail) },
				func(msg *data.TransferOrderDetail) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncLogWater:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.LogWater { return new(data.LogWater) },
				func(msg *data.LogWater) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncDetail:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.Detail { return new(data.Detail) },
				func(msg *data.Detail) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncLogLogin:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.LogLogin { return new(data.LogLogin) },
				func(msg *data.LogLogin) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncExternalBet:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.NsqLogExternalBet { return new(data.NsqLogExternalBet) },
				func(msg *data.NsqLogExternalBet) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncExternalReward:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.NsqLogExternalReward { return new(data.NsqLogExternalReward) },
				func(msg *data.NsqLogExternalReward) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncExternalCancel:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.NsqLogExternalCancel { return new(data.NsqLogExternalCancel) },
				func(msg *data.NsqLogExternalCancel) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncActivityBetRankPrize:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.ActivityBetRankPrize { return new(data.ActivityBetRankPrize) },
				func(msg *data.ActivityBetRankPrize) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncActivityTurnDrawLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.ActivityTurnDrawLog { return new(data.ActivityTurnDrawLog) },
				func(msg *data.ActivityTurnDrawLog) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncActivityTurnPrizeLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.ActivityTurnPrizeLog { return new(data.ActivityTurnPrizeLog) },
				func(msg *data.ActivityTurnPrizeLog) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncLaunchInviteLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.LaunchInviteLog { return new(data.LaunchInviteLog) },
				func(msg *data.LaunchInviteLog) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncShareAgentIncomeRecord:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.ShareAgentIncomeRecord { return new(data.ShareAgentIncomeRecord) },
				func(msg *data.ShareAgentIncomeRecord) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncShareAgentIncomeRecordTackLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.ShareAgentIncomeRecordTackLog { return new(data.ShareAgentIncomeRecordTackLog) },
				func(msg *data.ShareAgentIncomeRecordTackLog) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncLogVBDiamond:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.LogVbDiamond { return new(data.LogVbDiamond) },
				func(msg *data.LogVbDiamond) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncLogOutDiamond:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.LogOutDiamond { return new(data.LogOutDiamond) },
				func(msg *data.LogOutDiamond) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncVolatilitySubsidy:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.VolatilitySubsidy { return new(data.VolatilitySubsidy) },
				func(msg *data.VolatilitySubsidy) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncOnlineUsersLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.OnlineUsersLog { return new(data.OnlineUsersLog) },
				func(msg *data.OnlineUsersLog) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncUserCoupon:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.UserCoupon { return new(data.UserCoupon) },
				func(msg *data.UserCoupon) error {
					return ParseAndInsert([]any{msg})
				})
		case mq.TopicSyncCustomerChatSession:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *CustomerChatSession { return new(CustomerChatSession) },
				func(msg *CustomerChatSession) error {
					return InsertBatch([]any{msg})
				})
		case mq.TopicSyncCustomerChatMessage:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *CustomerChatMessage { return new(CustomerChatMessage) },
				func(msg *CustomerChatMessage) error {
					return InsertBatch([]any{msg})
				})
		case mq.TopicSyncCustomerSeatLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *CustomerSeatLog { return new(CustomerSeatLog) },
				func(msg *CustomerSeatLog) error {
					return InsertBatch([]any{msg})
				})
		case mq.TopicSyncCumRechargeWheelUnlockLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.CumRechargeWheelUnlockLog { return new(data.CumRechargeWheelUnlockLog) },
				func(msg *data.CumRechargeWheelUnlockLog) error {
					return InsertBatch([]any{msg})
				})
		case mq.TopicSyncCumRechargeWheelSpinLog:
			mq.NatsCreateConsumer(consumerName, mq.StreamSync, topic, func() *data.CumRechargeWheelSpinLog { return new(data.CumRechargeWheelSpinLog) },
				func(msg *data.CumRechargeWheelSpinLog) error {
					return InsertBatch([]any{msg})
				})
		}
	}
}
