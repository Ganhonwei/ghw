package report

import (
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
)

func initNatsConsumer() {
	consumerName := "report"

	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportRegister, func() *pb.EventRegister { return &pb.EventRegister{} }, handlerEventRegister)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportLogin, func() *pb.EventLogin { return &pb.EventLogin{} }, handlerEventLogin)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportDeposit, func() *pb.EventDeposit { return &pb.EventDeposit{} }, handlerEventDeposit)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportAdParam, func() *pb.EventAdParam { return &pb.EventAdParam{} }, handlerEventAdParam)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportFBRegister, func() *pb.FBEventRegister { return &pb.FBEventRegister{} }, handlerEventFBRegister)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportFBDeposit, func() *pb.FBEventDeposit { return &pb.FBEventDeposit{} }, handlerEventFBDeposit)
	mq.NatsCreateConsumerExtend(consumerName, mq.StreamReport, mq.TopicReportLog, func() *pb.NsqLog { return &pb.NsqLog{} }, handlerEventLog)
}
