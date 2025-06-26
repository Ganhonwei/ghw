package data

import (
	"gopkg.in/mgo.v2/bson"
)

// vip
type Vip struct {
	Id         int   `bson:"_id" json:"id"`                 // ID
	NextLevel  int   `json:"nextLevel" bson:"next_level"`   // 下一级
	DailySign  int64 `json:"dailySign" bson:"daily_sign"`   //VB每日签到
	WeeklySign int64 `json:"weeklySign" bson:"weekly_sign"` //VB每周签到
	Recharge   int64 `json:"recharge" bson:"recharge"`      // 累计充值
	// DailyReward   int64 `json:"dailyReward" bson:"daily_reward"`     // 每日领奖
	// WeeklyReward  int64 `json:"weeklyReward" bson:"weekly_reward"`   // 每周领奖
	// UpgradeRewrad int64 `json:"upgradeRewrad" bson:"upgrade_rewrad"` // 升级奖励
	WithdrawCount int `json:"withdrawCount" bson:"withdraw_count"` // 每日提现次数
	Commission    int `json:"commission" bson:"commission"`        // 佣金比例
	VoiceSwitch   int `json:"voiceSwitch" bson:"voice_switch"`     // 语音开关
	EmojiPrice    int `json:"emojiPrice" bson:"emoji_price"`       // 表情价格
	UnlockPhoto   int `json:"unlockPhoto" bson:"unlock_photo"`     // 解锁头像
}

func GetVipList() []Vip {
	var list []Vip
	ListByQ(Vips, nil, &list)
	return list
}

// Save 写入数据库
func (t *Vip) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(Vips, bson.M{"_id": t.Id}, t)
}

type UserVip struct {
	Init           bool    `json:"init" bson:"init"`                    //是否初始化
	BeforeLv       int     `json:"before_lv" bson:"before_lv"`          //之前等级
	Lv             int     `json:"lv" bson:"lv"`                        //当前等级
	MaxLv          int     `json:"max_lv" bson:"max_lv"`                //历史最高等级
	Daily          bool    `json:"daily" bson:"daily"`                  //每日奖励
	Weekly         bool    `json:"weekly" bson:"weekly"`                //每周奖励
	WeeklyTime     int64   `json:"weeklyTime" bson:"weekly_time"`       //下次领取周领奖时间
	Weekday        int     `json:"weekday" bson:"weekday"`              //周几
	LevelReward    []int32 `json:"levelReward" bson:"level_reward"`     //升级奖励
	Exp            int64   `json:"exp" bson:"exp"`                      //当前经验值
	WithdrawCount  int     `json:"withdrawCount" bson:"withdraw_count"` //每日提现次数
	Version        int     `json:"version" bson:"version"`              //版本
	ResetTime      int64   `json:"resetTime" bson:"reset_time"`
	WithdrawAmount int     `json:"withdrawAmount" bson:"withdraw_amount"` //每日提现金额
}

type VipBankLog struct {
	Date    int64 `json:"date" bson:"date"`       // 时间
	LogType int32 `json:"logType" bson:"logType"` // 类型
	Amount  int64 `json:"amount" bson:"amount"`   // 金额
}
