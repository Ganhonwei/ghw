package withdraw

import (
	"context"
	"errors"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

var AutoTransferSwitch, AutoDeliverySwitch bool

func start() {
	setting := data.GetSetting(bson.M{"stype": 4, "rtype": data.AUTOTRANSFER})
	if setting != nil {
		AutoTransferSwitch = setting.Status == 1
	}
	setting = data.GetSetting(bson.M{"stype": 4, "rtype": data.AUTODELIVERY})
	if setting != nil {
		AutoDeliverySwitch = setting.Status == 1
	}
	StartAutoTransferChecker()
}

func checkTransfer() error {
	// if !AutoTransferSwitch {
	// 	return nil
	// }
	err := AutoTimeoutTransfer()
	if err != nil {
		return err
	}
	err = AutoTimeoutTransferRapid()
	if err != nil {
		return err
	}
	return nil
}

func ResetTransferCash() error {
	transferCash := table.GetTables().PaymentTable.Get().WithdrawTransferReimburse
	_, err := myredis.Redis().Set(context.Background(), data.TransferCashKey, transferCash*100, 0).Result()
	if err != nil {
		glog.Infof("ResetTransferCash reset fail, err:%v", err)
		return err
	}
	glog.Infof("ResetTransferCash reset: %d, time: %s", transferCash, utils.Time2Str(time.Now()))
	return nil
}

// 超时转单
func AutoTimeoutTransfer() error {
	endTime := time.Now().UTC().AddDate(0, 0, -3)
	timeout := table.GetTables().PaymentTable.Get().WithdrawReviewTimeOut
	rapid := table.GetTables().PaymentTable.Get().WithdrawReviewTimeOutRapid
	transferTimes := table.GetTables().PaymentTable.Get().WithdrawTimeOutTimes
	if rapid > 0 {
		// 超时需转单，大于需转单时长，小于急需转单时长
		endTime = time.Now().UTC().Add(-time.Duration(rapid) * time.Minute)
	}
	if timeout == 0 {
		return nil
	}

	// 查询提单成功的订单(超过3天不查了)
	startTime := time.Now().UTC().Add(-time.Duration(timeout) * time.Minute)
	query := bson.M{
		"order_status": data.OrderSuccess,
		"repeat_times": 0,
		"e_time": bson.M{"$lte": startTime,
			"$gt": endTime,
		},
	}
	list := data.WithdrawRecordListByQ(query)
	tgMsg := make([]string, 0)

	transferCashStr, _ := myredis.Redis().Get(context.Background(), data.TransferCashKey).Result()
	transferCash, _ := strconv.ParseInt(transferCashStr, 10, 64)
	if transferCash < 0 {
		glog.Warningf("转单补贴金不足, 不进行转单")
		return errors.New("transferCash not enough")
	}

	// var ids []string
	for _, order := range list {
		if order.RepeatTimes > 0 {
			continue
		}
		var times int32
		exist := false
		for _, detail := range order.TransferDetail {
			if detail.Reason == data.TimeoutTransfer &&
				detail.TType == data.AutoTransfer {
				times++
			}
			// 如果还存在超时急需转单就不转此单
			if detail.Status == data.WithdrawUrgentTransfer {
				exist = true
				break
			}
		}
		if exist {
			glog.Warningf("用户%s 已存在转单记录, 不进行转单", order.Userid)
			continue
		}
		// 开了自动转单, 超时转单次数限制
		order.UpdateOrderStatus(data.WithdrawTimeoutTransfer)
		// 检查当前有没有转单补贴金
		if transferCash < int64(order.Amount) {
			glog.Warningf("转单补贴金不足, 不进行转单, orderId:%s", order.OrderID)
			continue
		}
		if AutoTransferSwitch && times < transferTimes {
			mq.NatsPublish(mq.TopicWithdrawTransfer, &pb.WithdrawTransfer{
				Ttype:   data.AutoTransfer,
				Reason:  int32(data.TimeoutTransfer),
				OrderId: order.OrderID,
				Name:    "System",
			})
			// 扣补贴金
			transferCash -= int64(order.Amount)
			glog.Infof("order:%s, transferCash:%d", order.OrderID, transferCash)
		}
		// tg通知
		tgMsg = append(tgMsg, order.OrderID)
	}
	myredis.Redis().Set(context.Background(), data.TransferCashKey, transferCash, 0)
	// 没开自动转单，更新状态
	// if len(ids) > 0 {
	// 	status := data.WithdrawTimeoutTransfer
	// 	data.UpdateAll(data.WithdrawRecords,
	// 		bson.M{"_id": bson.M{"$in": ids}},
	// 		bson.M{"$set": bson.M{"order_status": status}})
	// }
	// tg消息
	if len(tgMsg) > 0 {
		msg := &pb.AlertorMail{
			Subject: "超时转单",
			Message: strings.Join(tgMsg, "\n"),
		}
		mq.NatsPublish(mq.TopicAlertEmail, msg)
	}
	return nil
}

// 超时急转单
func AutoTimeoutTransferRapid() error {
	endTime := time.Now().UTC().AddDate(0, 0, -3)
	timeout := table.GetTables().PaymentTable.Get().WithdrawReviewTimeOutRapid
	transferTimes := table.GetTables().PaymentTable.Get().WithdrawTimeOutTimes
	if timeout == 0 {
		return nil
	}

	// 查询提单成功的订单(超过3天不查了)
	startTime := time.Now().UTC().Add(-time.Duration(timeout) * time.Minute)
	query := bson.M{
		"order_status": bson.M{"$in": []int32{data.OrderSuccess, data.WithdrawTimeoutTransfer}},
		"repeat_times": 0,
		"e_time": bson.M{"$lte": startTime,
			"$gt": endTime,
		},
	}
	list := data.WithdrawRecordListByQ(query)
	tgMsg := make([]string, 0)

	transferCashStr, _ := myredis.Redis().Get(context.Background(), data.TransferCashKey).Result()
	transferCash, _ := strconv.ParseInt(transferCashStr, 10, 64)
	if transferCash < 0 {
		glog.Warningf("转单补贴金不足, 不进行转单")
		return errors.New("transferCash not enough")
	}

	// var ids []string
	for _, order := range list {
		if order.RepeatTimes > 0 {
			continue
		}
		var times int32
		for _, detail := range order.TransferDetail {
			if detail.Reason == data.TimeoutUrgentTransfer &&
				detail.TType == data.AutoTransfer {
				times++
			}
		}
		// 开了自动转单, 超时转单次数限制,加配置
		order.UpdateOrderStatus(data.WithdrawUrgentTransfer)
		// 检查当前有没有转单补贴金
		if transferCash < int64(order.Amount) {
			glog.Warningf("转单补贴金不足, 不进行转单, orderId:%s", order.OrderID)
			continue
		}
		if AutoTransferSwitch && times < transferTimes {
			mq.NatsPublish(mq.TopicWithdrawTransfer, &pb.WithdrawTransfer{
				Ttype:   data.AutoTransfer,
				Reason:  int32(data.TimeoutUrgentTransfer),
				OrderId: order.OrderID,
				Name:    "System",
			})
			// 扣补贴金
			transferCash -= int64(order.Amount)
		}
		// tg通知
		tgMsg = append(tgMsg, order.OrderID)
	}
	myredis.Redis().Set(context.Background(), data.TransferCashKey, transferCash, 0)
	// 没开自动转单，更新状态
	// if len(ids) > 0 {
	// 	status := data.WithdrawTimeoutTransfer
	// 	data.UpdateAll(data.WithdrawRecords,
	// 		bson.M{"_id": bson.M{"$in": ids}},
	// 		bson.M{"$set": bson.M{"order_status": status}})
	// }
	// tg消息
	if len(tgMsg) > 0 {
		msg := &pb.AlertorMail{
			Subject: "超时急转单",
			Message: strings.Join(tgMsg, "\n"),
		}
		mq.NatsPublish(mq.TopicAlertEmail, msg)
	}
	return nil
}

// 失败自动转单
func AutoFailTransfer(order *data.WithdrawRecord) {
	if !AutoTransferSwitch {
		return
	}

	times := table.GetTables().PaymentTable.Get().FailTransferTimes
	details := order.TransferDetail
	if len(details)-1 >= int(times) {
		// 超过以转单次数限制
		return
	}
	// 转单
	mq.NatsPublish(mq.TopicWithdrawTransfer, &pb.WithdrawTransfer{
		Ttype:   data.AutoTransfer,
		Reason:  int32(data.FailedTransfer),
		OrderId: order.OrderID,
		Name:    "System",
	})
}

func deliveryOrder() {
	query := bson.M{
		"order_status": data.WaitDelivery,
	}
	startTime := time.Now().Unix()
	var success, fail int
	list := data.WithdrawRecordListByQ(query)
	for _, order := range list {
		// 自动派单没开
		if !AutoDeliverySwitch {
			break
		}
		err := mq.NatsPublish(mq.TopicWithdrawTransfer, &pb.WithdrawTransfer{
			Ttype:   data.NoTransfer,
			Reason:  int32(data.NormalTransfer),
			OrderId: order.OrderID,
			Name:    "System",
		})
		if err != nil {
			fail++
			glog.Errorf("deliveryOrder publish fail, orderId:%s, err:%v", order.OrderID, err)
			continue
		}
		success++
		glog.Infof("deliveryOrder publish success, orderId:%s, success:%d, fail:%d", order.OrderID, success, fail)
	}
	glog.Infof("deliveryOrder end, success:%d, fail:%d, time:%d", success, fail, time.Now().Unix()-startTime)
}
