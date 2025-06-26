package online

import (
	"goserver/pkg/data/mq"
)

func initNatsRouter() (err error) {
	initNatsConsumer()

	// 请求在线用户人数
	mq.NatsSubscribe(mq.RequestOnlineUserCount, func() *mq.RequestEmptyArgs { return &mq.RequestEmptyArgs{} },
		requestOnlineUserMinuteReq)

	// 请求在线用户列表
	mq.NatsSubscribe(mq.RequestOnlineUserList, func() *mq.RequestEmptyArgs { return &mq.RequestEmptyArgs{} },
		requestOnlineUserListReq)

	return
}
