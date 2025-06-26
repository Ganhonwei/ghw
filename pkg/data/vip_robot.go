package data

import "github.com/globalsign/mgo/bson"

// 生成人机的VIP规则配置
type VipRobot struct {
	Id         string `bson:"_id" json:"id"`                // id
	VipChoices []int  `json:"vipChoices" bson:"vipChoices"` // 人机VIP范围
	VipWeights []int  `json:"vipWeights" bson:"vipWeights"` // 人机VIP权重
}

func (t *VipRobot) Save() bool {
	return Upsert(VipRobots, bson.M{"_id": t.Id}, t)
}

func GetVipRobots() []*VipRobot {
	var list []*VipRobot
	q := bson.M{}
	ListByQ(VipRobots, q, &list)
	return list
}
