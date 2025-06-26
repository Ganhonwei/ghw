package data

import "github.com/globalsign/mgo/bson"

// 新提现跑马灯
type MarqueeWithdraw struct {
	Id             int    `bson:"_id" json:"id"`
	TimeRange      string `bson:"time_range" json:"timeRange"`           // `xls:"触发时间段"`
	Seconds        []int  `bson:"seconds" json:"seconds"`                // `xls:"随机秒数"`
	Withdraws      []int  `bson:"withdraws" json:"withdraws"`            // `xls:"提现档位范围"`
	WithdrawWeight []int  `bson:"withdraw_weight" json:"withdrawWeight"` // `xls:"提现档位权重"`
	TimeRanges     [2]int `bson:"-" json:"-"`                            // 触发时间秒
}

// Save 写入数据库
func (t *MarqueeWithdraw) Save() bool {
	return Upsert(MarqueeWithdraws, bson.M{"_id": t.Id}, t)
}

func GetMarqueeWithdrawList() []*MarqueeWithdraw {
	var list []*MarqueeWithdraw
	ListByQ(MarqueeWithdraws, nil, &list)
	return list
}
