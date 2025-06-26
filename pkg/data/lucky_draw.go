package data

import (
	"goserver/pkg/glog"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// 小米14活动奖励配置
type LuckyDrawReward struct {
	Id     int   `bson:"_id" json:"id"`        // id 奖励等级
	Reward int64 `json:"reward" bson:"reward"` // 奖励彩金
	Limit  int32 `json:"limit" bson:"limit"`   // 中奖人数
	Rate   int32 `json:"rate" bson:"rate"`     // 每张号码中奖率（万分比）
}

// Save 写入数据库
func (t *LuckyDrawReward) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(LuckyDrawRewards, bson.M{"_id": t.Id}, t)
}

func GetLuckyDrawRewardList() []*LuckyDrawReward {
	var list []*LuckyDrawReward
	ListByQ(LuckyDrawRewards, nil, &list)
	return list
}

// 小米14活动自动放号配置
type LuckyDrawRobot struct {
	Id          string  `bson:"_id" json:"id"`                  // id
	Counts      []int32 `bson:"counts" json:"counts"`           // `xls:"已放号数量"`
	TimePeriod1 []int32 `bson:"timePeriod1" json:"timePeriod1"` // `xls:"2:00:00-6:59:59"`
	TimePeriod2 []int32 `bson:"timePeriod2" json:"timePeriod2"` // `xls:"7:00:00-12:59:59"`
	TimePeriod3 []int32 `bson:"timePeriod3" json:"timePeriod3"` // `xls:"13:00:00-18:59:59"`
	TimePeriod4 []int32 `bson:"timePeriod4" json:"timePeriod4"` // `xls:"19:00:00-1:59:59"`
}

// Save 写入数据库
func (t *LuckyDrawRobot) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(LuckyDrawRobots, bson.M{"_id": t.Id}, t)
}

func GetLuckyDrawRobotList() []*LuckyDrawRobot {
	var list []*LuckyDrawRobot
	ListByQ(LuckyDrawRobots, nil, &list)
	return list
}

// 小米活动奖券发放记录
type LuckyDrawNumber struct {
	Id     string    `bson:"_id" json:"id"`        // unique ID
	Round  int64     `bson:"round" json:"round"`   // 轮数
	Number int32     `bson:"number" json:"number"` // 抽奖号码
	Userid string    `bson:"userid" json:"userid"` //中奖人id
	Robot  bool      `bson:"robot" json:"robot"`   // 是否机器人
	Ctime  time.Time `bson:"ctime" json:"ctime"`   //创建时间
}

func (t *LuckyDrawNumber) Save() bool {
	return Insert(LuckDrawNumbers, t)
}

func GetLuckyDrawCurRound() (curRound int64, err error) {
	number := &LuckyDrawNumber{}
	err = LuckDrawNumbers.Find(bson.M{}).Sort("-round").Limit(1).One(number)
	if err != nil {
		if err == mgo.ErrNotFound {
			return 0, nil
		}
		return
	}
	if number.Id == "" {
		return
	}
	curRound = number.Round
	return
}

func GetDrawNumbersByRound(round int64) (numbers []*LuckyDrawNumber) {
	ListByQ(LuckDrawNumbers, bson.M{"round": round}, &numbers)
	return
}

func GetDrawNumbersByRoundMN(roundM, roundN int64) (numbers []*LuckyDrawNumber) {
	m := bson.M{"round": bson.M{"$gte": roundM, "$lte": roundN}}
	err := LuckDrawNumbers.Find(m).Sort("round").All(&numbers)
	if err != nil {
		glog.Error("GetDrawGivesByRoundMN err: %d,%d, %v", roundM, roundN, err)
	}
	return
}

// 查询当前回合最大中奖号码
func GetRoundDrawNumberMax(round int64, notNumbers ...int32) (number *LuckyDrawNumber, err error) {
	m := bson.M{
		"round":  round,
		"number": bson.M{"$nin": notNumbers},
	}
	err = LuckDrawNumbers.Find(m).Sort("-number").Limit(1).One(number)
	return
}

type LuckyDrawGive struct {
	Id       string    `bson:"_id" json:"id"`              // unique ID
	Round    int64     `bson:"round" json:"round"`         // 中奖轮数，第几期
	RewordId int32     `bson:"reward_id" json:"reward_id"` // 几等奖
	Reword   int64     `bson:"reword" json:"reword"`       // 中奖彩金
	Userid   string    `bson:"userid" json:"userid"`       // 中奖人id
	Number   string    `bson:"number" json:"number"`       // 中奖号码
	Robot    bool      `bson:"robot" json:"robot"`         // 是否机器人
	Taked    bool      `bson:"taked" json:"taked"`         // 已领取
	Ctime    time.Time `bson:"ctime" json:"ctime"`         // 创建时间
}

func (t *LuckyDrawGive) Save() bool {
	return Insert(LuckyDrawGives, t)
}

func (t *LuckyDrawGive) UpdateTaked() {
	ok := Update(LuckyDrawGives, bson.M{"_id": t.Id}, bson.M{"$set": bson.M{"taked": t.Taked}})
	if !ok {
		glog.Errorf("UpdateTaked err: %v, %v", t.Id, t.Taked)
	}
}

func GetDrawGivesByRound(round int64) (gives []*LuckyDrawGive) {
	ListByQ(LuckyDrawGives, bson.M{"round": round}, &gives)
	return
}

func GetDrawGivesByRoundMN(roundM, roundN int64) (gives []*LuckyDrawGive) {
	m := bson.M{"round": bson.M{"$gte": roundM, "$lte": roundN}}
	err := LuckyDrawGives.Find(m).Sort("round").All(&gives)
	if err != nil {
		glog.Error("GetDrawGivesByRoundMN err: %d,%d, %v", roundM, roundN, err)
	}
	return
}
