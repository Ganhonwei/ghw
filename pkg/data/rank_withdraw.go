package data

import (
	"github.com/globalsign/mgo/bson"
)

// 提现排行榜配置
type RankWithdraw struct {
	Id           string `bson:"_id" json:"id"`                     // id
	RobotType    int32  `json:"robot_type" bson:"robot_type"`      // 人机组 0固定1随机
	RankMinute   int32  `json:"rankMinute" bson:"rank_minute"`     // 每x分钟触发提现判断
	WithdrawRate int32  `json:"withdrawRate" bson:"withdraw_rate"` // 提现概率(万分比）
	Choices      []int  `json:"choices" bson:"choices"`            // 提现档位
	VipChoices   []int  `json:"vipChoices" bson:"vipChoices"`      // 人机VIP范围
	VipWeights   []int  `json:"vipWeights" bson:"vipWeights"`      // 人机VIP权重
}

func (t *RankWithdraw) Save() bool {
	return Upsert(RankWithdraws, bson.M{"_id": t.Id}, t)
}

func GetRankWithdraw() []*RankWithdraw {
	var list []*RankWithdraw
	q := bson.M{}
	ListByQ(RankWithdraws, q, &list)
	return list
}

// 提现排行mongo
// type RankWithdrawList struct {
// 	Date  string             `json:"date" bson:"date"` // 日期，周排行为一周六日期
// 	Type  int32              `json:"type" bson:"type"` // 1:日排行 2:周排行
// 	Ranks []RankWithdrawItem `json:"ranks" bson:"ranks"`
// }

// type RankWithdrawItem struct {
// 	Robot    bool   `json:"robot,omitempty" bson:"robot"`
// 	Userid   string `json:"userid,omitempty" bson:"userid"`
// 	Username string `json:"username,omitempty" bson:"username"`
// 	Avatar   string `json:"avatar,omitempty" bson:"avatar"`
// 	Amount   int32  `json:"amount,omitempty" bson:"amount"`
// }
