package tasks

import (
	"errors"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/glog"

	"github.com/hibiken/asynq"
)

var redisAddr string

func InitAsynqClient(addr string) error {
	redisAddr = addr
	return nil
}

func CreateWithdrawSchedule(order *entity.WithdrawOrder, retryTimes int) error {
	// client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	// if client == nil {
	// 	return errors.New("asynq client is nil")
	// }
	// defer client.Close()

	// // TODO 没开自动转单不处理
	// if false {
	// 	return nil
	// }
	// // 状态, 只有失败转单状态才能转
	// if order.OrderStatus != data.WithdrawFailTransfer {
	// 	return nil
	// }
	// // 判断转单次数

	// // order.OrderStatus = data.WithdrawFail
	// order.InspectTimes++
	// defer service.UpdateWithDrawOrder(order)

	// retryTimes = order.InspectTimes
	// task, err := NewBadWithdrawDeliveryTask(order.OrderID, retryTimes)
	// if err != nil {
	// 	return err
	// }
	// // 给个3s的延迟
	// delay := 3 * time.Second
	// // switch retryTimes {
	// // case 1, 3:
	// // case 2:
	// // 	glog.Infof("%s inspect %d times,need change channel", order.OrderID, retryTimes)
	// // default:
	// // 	// glog.Infof("%s inspect %d times,need change channel", order.OrderID, retryTimes)
	// // 	return nil
	// // }
	// info, err := client.Enqueue(task, asynq.ProcessIn(delay))
	// if err != nil {
	// 	glog.Errorf("task fail, id:%s, err:%s", order.OrderID, err)
	// 	return err
	// }

	// // order.InspectTimes++
	// glog.Infof("enqueued task: id=%s queue=%s order:%s", info.ID, info.Queue, order.OrderID)
	return nil
}

func CreatePayLogSchedule(orderid, userid string, ptype int, channelId uint32, amount int64) error {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	if client == nil {
		return errors.New("asynq client is nil")
	}
	defer client.Close()

	task, err := NewPayLogDeliveryTask(orderid, userid, ptype, channelId, amount)
	if err != nil {
		return err
	}
	info, err := client.Enqueue(task)
	if err != nil {
		glog.Errorf("task fail, id:%s, err:%s", orderid, err)
		return err
	}

	glog.Infof("enqueued task: pay log id=%s queue=%s order:%s", info.ID, info.Queue, orderid)
	return nil
}
