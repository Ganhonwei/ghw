package pay

import (
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"gopkg.in/mgo.v2/bson"
)

// 开启查单
func InspectOrder() []*entity.PayOrder {
	now := utils.BsonNow().UnixMilli()
	list := make([]*entity.PayOrder, 0)
	exception := make([]*entity.PayOrder, 0)
	//查询10-20分钟前的订单.
	service.ListByQ(service.PayOrders, bson.M{"order_status": 1, "ctime": bson.M{"$lt": now - 600*1000, "$gte": now - 1200*1000}}, &list)
	if list == nil {
		glog.Error("未查询到订单:", utils.BsonNow())
		return nil
	}
	for _, order := range list {
		// 查单
		if !inspectOrderHandler(order) && order.PayWay != 1 {
			exception = append(exception, order)
		}
	}
	return exception
}

// 向第三方核查订单
func inspectOrderHandler(order *entity.PayOrder) bool {
	ment := service.GetChannelMent(order.ChannelId)
	if ment == nil {
		return false
	}
	rsp, err := ment.InspectSubmit(order)
	if err != nil {
		glog.Errorf("submit inspect request fail, order:%s, channel:%d", order.OrderID, order.ChannelId)
		return false
	}

	err1 := ment.InspectResponse(rsp, order)
	if err1 != nil {
		glog.Infof("submit inspect request fail, order:%s, channel:%d", order.OrderID, order.ChannelId)
		return false
	}

	// 支付成功的订单，需要手动回调
	notify := new(data.PayNotify)
	notify.OrderID = order.OrderID
	notify.MerOrderID = order.MerOrderID
	notify.BusinessID = order.BusinessID
	notify.Status = 2
	notify.Amount = int(order.Amount)
	notify.Timestamp = utils.BsonNow().UnixMilli()
	notify.Repair = true
	msg, err2 := service.PayNotify(notify, 1)
	if err2 != nil {
		return false
	}
	if string(msg) != "0" {
		glog.Errorf("callback fail, err:%s", msg)
		return false
	}
	order.OrderStatus = 2
	service.UpdateOrder(order)
	return true
}

// 提现核单
func InspectWithdrawHandler(order *entity.WithdrawOrder) {

}
