package report

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func handlerEventLog(log *pb.NsqLog, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive log: %v", log)

	switch log.Typ {
	case pb.ExternalBet:
		body := &pb.NsqLogExternalBet{}
		err = body.Unmarshal(log.Body)
		if err != nil {
			ack = false
			return
		}
		data.LogExternalBet(body)
	case pb.ExternalReward:
		body := &pb.NsqLogExternalReward{}
		err = body.Unmarshal(log.Body)
		if err != nil {
			ack = false
			return
		}
		data.LogExternalReward(body)
	case pb.ExternalCancel:
		body := &pb.NsqLogExternalCancel{}
		err = body.Unmarshal(log.Body)
		if err != nil {
			ack = false
			return
		}
		data.LogExternalCancel(body)
	case pb.EfiTransactionLog:
		body := &pb.EfiTransactionReq{}
		err = body.Unmarshal(log.Body)
		if err != nil {
			ack = false
			return
		}
		data.LogEfiTransaction(body)
	}

	ack = true
	return
}
