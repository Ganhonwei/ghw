package service

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/utils"
	"time"
)

func InitNatsRouter() {
	consumerName := "web-admin"
	// 玩家已读消息订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameCustomerRead, func() *pb.PublishGameCustomerRead { return new(pb.PublishGameCustomerRead) },
		func(msg *pb.PublishGameCustomerRead) (err error) {
			err = MessageService.CustomerMsgRead(msg.Userid, msg.Msgs)
			return
		})

	// 客服会话玩家评分
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameCustomerScore, func() *pb.CustomerScoreReq { return new(pb.CustomerScoreReq) },
		func(msg *pb.CustomerScoreReq) (err error) {
			err = MessageService.CustomerSessionScore(msg.Userid, msg.SessionId, msg.Score)
			return
		})

	// 客服发送消息请求
	mq.NatsSubscribe(mq.RequestCustomerSend, func() *pb.CustomerSendReq { return new(pb.CustomerSendReq) },
		requestCustomerSend)
}

// 客服发送消息请求 handler
func requestCustomerSend(req *pb.CustomerSendReq) (ret *pb.CustomerMessage, err error) {
	ret = new(pb.CustomerMessage)
	msg, err := MessageService.CustomerSend(req.Userid, req.Username, req.Ctype, req.Content, req.Filename, req.Filesize)
	if err != nil {
		return
	}
	ret.MessageId = fmt.Sprint(msg.Id)
	ret.Reply = msg.Reply
	ret.Ctype = msg.Ctype
	ret.Content = msg.Content
	ret.Ctime = time.UnixMilli(msg.Ctime).In(Location()).Format(utils.FORMAT)
	ret.Read = msg.Read
	return
}
