package betrank

import (
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
)

func initNatsRouter() (err error) {
	err = initNatsConsumer()
	if err != nil {
		return
	}

	// 排行榜拉取nats请求
	mq.NatsSubscribe(mq.RequestActivityBetRankList, func() *mq.RequestActivityBetRankListArgs { return &mq.RequestActivityBetRankListArgs{} },
		requestActivityBetRankListReq)

	// 弹窗 skip for today 请求
	mq.NatsSubscribe(mq.RequestActivityBetRankSkipWindow, func() *mq.RequestActivityBetRankSkipWindowArgs { return &mq.RequestActivityBetRankSkipWindowArgs{} },
		requestActivityBetRankSkipWindow)

	// 排行榜更新请求
	mq.NatsSubscribe(mq.RequestActivityBetRankUpdate, func() *mq.RequestActivityBetRankUpdateArgs { return &mq.RequestActivityBetRankUpdateArgs{} },
		requestActivityBetRankUpdate)

	// 领奖
	mq.NatsSubscribe(mq.RequestActivityBetRankReword, func() *mq.RequestActivityBetRankRewordArgs { return &mq.RequestActivityBetRankRewordArgs{} },
		requestActivityBetRankReword)

	// 个人领取记录请求
	mq.NatsSubscribe(mq.RequestActivityBetRankMyRecord, func() *mq.RequestUserArgs { return &mq.RequestUserArgs{} },
		requestActivityBetRankMyRecord)

	// 排行榜历史请求
	mq.NatsSubscribe(mq.RequestActivityBetRankHistory, func() *mq.RequestUserArgs { return &mq.RequestUserArgs{} },
		requestActivityBetRankHistory)

	return
}

func initNatsConsumer() (err error) {
	consumerName := "activity-betrank"
	// 打码上报订阅
	inst_GameBets := func() *pb.PublishGameBets { return &pb.PublishGameBets{} }
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameBets, inst_GameBets, handlePublishBets,
		mq.WithConsumerNoneAck(), mq.WithConsumerMaxDeliver(1))

	// 返回大厅订阅
	inst_GameToLobby := func() *pb.PublishGameToLobby { return &pb.PublishGameToLobby{} }
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameToLobby, inst_GameToLobby, func(data *pb.PublishGameToLobby) (err error) {
		// glog.Infof("to lobby: %s", data.Userid)
		// 判断弹窗
		ntfs, ntf, err := handleUserWindows(data.Userid)
		if err != nil {
			glog.Error(err)
			return
		}
		if ntf {
			mq.NatsPublish(mq.TopicActivityBetRankWindow, ntfs)
		}
		return
	})
	return
}
