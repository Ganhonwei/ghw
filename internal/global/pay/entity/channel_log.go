package entity

// 支付渠道日志
type PayChannelLog struct {
	Id             string  `bson:"_id"`                                   // 主键id
	Date           int64   `bson:"date"`                                  // 日期
	OrderId        string  `json:"orderId" bson:"order_id"`               // 订单ID
	PayChannel     uint32  `json:"payChannel" bson:"pay_channel"`         // 支付渠道
	PType          int32   `json:"pType" bson:"p_type"`                   // 支付类型 1：代收;2:代付
	UserId         string  `json:"userId" bson:"user_id"`                 // 玩家ID
	Amount         int64   `json:"amount" bson:"amount"`                  // 金额
	Rate           float64 `json:"rate" bson:"rate"`                      // 税率
	ActualAmount   int64   `json:"actualAmount" bson:"actual_amount"`     // 实际金额
	HandlingCharge int64   `json:"handlingCharge" bson:"handling_charge"` // 税费
	Balance        int64   `json:"balance" bson:"balance"`                // 实时余额
}
