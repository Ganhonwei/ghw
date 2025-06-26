package withdraw

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"

	"github.com/globalsign/mgo/bson"
)

func initNatsConsumer() {
	consumerName := "withdraw"

	mq.NatsCreateConsumer(consumerName, mq.StreamWithdraw, mq.TopicWithdrawAutoTransfer, func() *pb.WithdrawAutoTransfer { return &pb.WithdrawAutoTransfer{} },
		func(d *pb.WithdrawAutoTransfer) (err error) {
			setting := data.GetSetting(bson.M{"stype": 4, "rtype": data.AUTOTRANSFER})
			if setting != nil {
				AutoTransferSwitch = setting.Status == 1
			}
			setting = data.GetSetting(bson.M{"stype": 4, "rtype": data.AUTODELIVERY})
			if setting != nil {
				AutoDeliverySwitch = setting.Status == 1
			}
			if AutoDeliverySwitch {
				// 把代派单的都派出去
				deliveryOrder()
			}
			return
		})

	mq.NatsCreateConsumer(consumerName, mq.StreamWithdraw, mq.TopicWithdrawFail, func() *pb.WithdrawTransfer { return &pb.WithdrawTransfer{} },
		func(d *pb.WithdrawTransfer) (err error) {
			order := &data.WithdrawRecord{
				OrderID: d.OrderId,
			}
			order.Get()
			switch order.OrderStatus {
			case data.WithdrawFailTransfer, data.DeliveryFailTransfer:
				AutoFailTransfer(order)
			}
			return
		})
}
