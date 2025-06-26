package entity

import "time"

// 用户动作
type Action struct {
	Id         string    `bson:"_id"`
	Action     string    `bson:"action"`      // 动作类型
	Actor      string    `bson:"actor"`       // 操作角色
	ObjectType string    `bson:"object_type"` // 操作对象类型
	ObjectId   string    `bson:"object_id"`   // 操作对象id
	Extra      string    `bson:"extra"`       // 额外信息
	CreateTime time.Time `bson:"create_time"` // 更新时间
	Message    string    `bson:"message"`     // 格式化后的消息
}

// 金银铜卡
type WeekCard struct {
	Id              string    `bson:"_id" json:"id"`                              // id
	ReceiveTimes    int32     `bson:"receive_times" json:"receive_times"`         // 已领取次数
	LastReceiveTime time.Time `bson:"last_receive_time" json:"last_receive_time"` // 上次领取时间
	CanGet          int32     `bson:"can_get" json:"can_get"`                     // 可领取次数
}
