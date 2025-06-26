package trade

import (
	"goserver/internal/global/pay/entity"
	"goserver/internal/global/pay/service"
	"goserver/pkg/data"
)

func WithdrawOrderSubmit(order *entity.WithdrawOrder, rsp *data.WithdrawResponse) {
	defer service.UpdateWithDrawOrder(order)
	withdrawMent := service.GetChannelMent(order.ChannelId)
	if withdrawMent == nil {
		rsp.Code = 100
		rsp.Msg = "get withdraw fail"
		return
	}

	res, err := withdrawMent.WithdrawSubmit(order)
	// 更新订单信息
	order.RequestMsg = string(res)
	if err != nil {
		msgstr := "代付订单请求下单失败！,msg:" + err.Error()
		rsp.Code = 100
		rsp.Msg = msgstr
		return
	}

	isOrder, msg := withdrawMent.WithdrawSubmitResponse(res, order)
	if isOrder {
		rsp.Code = 200
	} else {
		rsp.Code = 100
		rsp.Msg = msg
	}
}
