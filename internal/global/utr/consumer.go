package utr

import (
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
)

func initNatsConsumer() {
	consumerName := "utr"

	mq.NatsCreateConsumerExtend(consumerName, mq.StreamUtr, mq.TopicUtrUpload, func() *pb.UploadUtrReq { return &pb.UploadUtrReq{} }, handlerEventUploadUtr)
}
