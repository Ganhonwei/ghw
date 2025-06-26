package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/internal/global/pay/trade"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/hibiken/asynq"
	"gopkg.in/mgo.v2/bson"
)

const (
	TypeWithdrawBadDelivery = "badwithdraw:deliver"
	TypePayLogDelivery      = "paylog:deliver"
)

type BadWithdrawDeliveryPayload struct {
	OrderId    string // 订单id
	RetryTimes int    // 重试次数
}

type PayLogDeliveryPayload struct {
	OrderId   string // 订单id
	Ptype     int    // 1.充值 2.提现
	ChannelId uint32 // 渠道
	Amount    int64  // 金额
	Userid    string // userid
}

func NewBadWithdrawDeliveryTask(orderId string, retryTimes int) (*asynq.Task, error) {
	payload, err := json.Marshal(BadWithdrawDeliveryPayload{OrderId: orderId, RetryTimes: retryTimes})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWithdrawBadDelivery, payload), nil
}

func NewPayLogDeliveryTask(orderId, userid string, ptype int, channelId uint32, amount int64) (*asynq.Task, error) {
	payload, err := json.Marshal(PayLogDeliveryPayload{OrderId: orderId, Ptype: ptype, ChannelId: channelId, Amount: amount, Userid: userid})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypePayLogDelivery, payload), nil
}

func HandleBadWithdrawDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p BadWithdrawDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	glog.Infof("Sending withdraw order: orderid=%s, retryTimes=%d", p.OrderId, p.RetryTimes)
	// bad withdraw order delivery code
	order, err := service.GetWithDrawOrder(p.OrderId)
	if err != nil {
		return err
	}
	if order.OrderStatus != data.WithdrawFail {
		glog.Error("withdraw order status fail")
		return nil
	}

	defer service.UpdateWithDrawOrder(order)

	order.UsedChannelId = append(order.UsedChannelId, strconv.Itoa(int(order.ChannelId)))
	// channelId := service.GetChannelId(false, order.UsedChannelId, order.PayWay)
	channelId := service.GetWithdrawChannelId(order.UsedChannelId, order.PayWay, int64(order.Amount), order.Country, order.Bank)
	if channelId != 0 {
		// 有可用渠道才切
		order.ChannelId = channelId
	} else {
		// TODO
		glog.Error("no access, order:", order.OrderID)
		CallbackWithdrawOrder(order, data.WithdrawFail)
		return nil
	}
	// 下单
	rsp := &data.WithdrawResponse{
		OrderID:    order.OrderID,
		ChannelID:  order.ChannelId,
		MerOrderID: order.MerOrderID,
	}
	trade.WithdrawOrderSubmit(order, rsp)
	if rsp.Code != 200 {
		// 轮询去
		CreateWithdrawSchedule(order, p.RetryTimes+1)
		return nil
	}

	// 更新订单信息 // TODO
	CallbackWithdrawOrder(order, data.OrderSuccess)
	return nil
	// // 第二次的时候需要换个渠道
	// if p.RetryTimes == 2 {
	// 	order.UsedChannelId = append(order.UsedChannelId, strconv.Itoa(int(order.ChannelId)))
	// 	// channelId := service.GetChannelId(false, order.UsedChannelId, order.PayWay)
	// 	channelId := service.GetWithdrawChannelId(order.UsedChannelId, order.PayWay, int64(order.Amount), order.Country, order.Bank)
	// 	if channelId != 0 {
	// 		// 有可用渠道才切
	// 		order.ChannelId = channelId
	// 	} else {
	// 		// TODO
	// 		glog.Error("no access, order:", order.OrderID)
	// 		CallbackWithdrawOrder(order, checkOrderStatus(order))
	// 		return nil
	// 	}
	// 	// 下单
	// 	rsp := &data.WithdrawResponse{
	// 		OrderID:    order.OrderID,
	// 		ChannelID:  order.ChannelId,
	// 		MerOrderID: order.MerOrderID,
	// 	}
	// 	trade.WithdrawOrderSubmit(order, rsp)
	// 	if rsp.Code != 200 {
	// 		// 轮询去
	// 		CreateWithdrawSchedule(order, p.RetryTimes+1)
	// 		return nil
	// 	}

	// 	// 更新订单信息 // TODO
	// 	CallbackWithdrawOrder(order, data.OrderSuccess)
	// 	return nil
	// }

	// retry := false
	// status := 3 // 默认状态是失败的
	// // 先查单
	// ment := service.GetChannelMent(order.ChannelId)
	// if ment != nil {
	// 	resp, err := ment.InspectWithdrawSubmit(order)
	// 	if err == nil {
	// 		status = ment.InspectWdResponse(resp, order)
	// 	}
	// }

	// if status == 2 {
	// 	// 掉单了，需要把状态改了,回调一下
	// 	_, err2 := CallbackWithdrawOrder(order, data.WithdrawSuccess)
	// 	if err2 != nil {
	// 		retry = true
	// 	} else {
	// 		order.OrderStatus = data.WithdrawSuccess
	// 	}
	// } else if status == 1 || status == 0 {
	// 	// 还在处理,需要改为提单成功
	// 	order.OrderStatus = data.Withdrawing
	// 	return nil
	// } else {
	// 	retry = true
	// }

	// if retry {
	// 	// 重试了3次之后还是失败就判定这单失败了，需要退钱
	// 	if p.RetryTimes >= 3 {
	// 		CallbackWithdrawOrder(order, checkOrderStatus(order))
	// 		return nil
	// 	}
	// 	CreateWithdrawSchedule(order, p.RetryTimes+1)
	// }
	// return nil
}

func checkOrderStatus(order *entity.WithdrawOrder) int {
	// switch order.ChannelId {
	// case 3011:
	// 	return mlpay.Config.GetWithdrawStatus(order)
	// }
	return data.WithdrawFail // 失败
}

func CallbackWithdrawOrder(order *entity.WithdrawOrder, status int) ([]byte, error) {
	ntf := new(data.PayNotify)
	ntf.OrderID = order.OrderID
	ntf.MerOrderID = order.MerOrderID
	ntf.BusinessID = ""
	ntf.Status = status
	ntf.Amount = int(order.Amount * 100)
	ntf.ChannelId = order.ChannelId
	ntf.Timestamp = utils.BsonNow().UnixMilli()
	ntf.Reason = data.FailedTransfer
	return service.PayNotify(ntf, 2)
}

// 成功订单日志
func HandlePayLogDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p PayLogDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	glog.Infof("Sending order log: orderid=%s, Ptype=%d", p.OrderId, p.Ptype)
	// bad withdraw order delivery code

	channel := service.GetPayChannel(utils.String(p.ChannelId))
	if channel.Id == "" {
		glog.Infof("channel not found or channel unopen: orderid=%s, Ptype=%d, channelId:%d", p.OrderId, p.Ptype, p.ChannelId)
		return nil
	}

	var actualAmount, handlingCharge int64
	if p.Ptype == 1 {
		actualAmount = int64(float64(p.Amount) * (1 - channel.PayRate/100))
		handlingCharge = int64(float64(p.Amount) * (channel.PayRate / 100))
	} else {
		// actualAmount = int64(float64(p.Amount)*(1-channel.WithdrawRate/100)) - channel.WithdrawFee
		handlingCharge = int64(float64(p.Amount)*(channel.WithdrawRate/100)) + channel.WithdrawFee
		actualAmount = p.Amount + handlingCharge
	}

	log := entity.PayChannelLog{
		Id:             bson.NewObjectId().String(),
		Date:           utils.BsonNow().UnixMilli(),
		OrderId:        p.OrderId,
		PayChannel:     p.ChannelId,
		PType:          int32(p.Ptype),
		UserId:         p.Userid,
		Amount:         p.Amount,
		Rate:           channel.PayRate,
		ActualAmount:   actualAmount,
		HandlingCharge: handlingCharge,
	}

	if p.Ptype == 2 {
		log.Rate = channel.WithdrawRate
	}

	// 查余额
	payment := service.GetChannelMent(p.ChannelId)
	if payment == nil {
		glog.Errorf("payment not found: orderid=%s, Ptype=%d, channelId:%d", p.OrderId, p.Ptype, p.ChannelId)
		log.Balance = 0
	} else {
		log.Balance = payment.BalanceQuery()
	}
	service.SavePayLog(log)

	// 本地渠道结算
	if p.Ptype == 1 {
		// 代收要加钱
		channel.Balance += int(actualAmount)
		service.Update(service.PayChannels, bson.M{"_id": channel.Id}, bson.M{"$inc": bson.M{"balance": actualAmount}})
	} else {
		// 代付要扣钱
		service.Update(service.PayChannels, bson.M{"_id": channel.Id}, bson.M{"$inc": bson.M{"balance": -actualAmount}})
		channel.Balance -= int(actualAmount)
	}

	// 预警
	if log.Balance/100 > 100000 && p.Ptype == 1 && service.Env != "dev" {
		url := "http://13.126.251.84:3001/api/push/Gho0T85mVf?msg=" + fmt.Sprintf("%s余额：%.2f", service.ChannelNameMap[int(p.ChannelId)], float64(log.Balance)/100)
		doGetHttp(url)
		// mail := &pb.AlertorMail{Subject: "渠道余额:", Message: fmt.Sprintf("%s：%.2f", service.ChannelNameMap[int(p.ChannelId)], float64(log.Balance)/100)}
		// body, _ := proto.Marshal(mail)
		// service.Producer.Publish(data.TopicAlertor, body)
	}
	return nil
}

func doGetHttp(targetUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	return respData, nil
}
