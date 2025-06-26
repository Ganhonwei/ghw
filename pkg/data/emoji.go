package data

import "gopkg.in/mgo.v2/bson"

// 人机表情
type Emoji struct {
	Id              int32         `json:"_id" bson:"_id"`
	RobotTypeWeight []int         `json:"robot_type_weight" bson:"robot_type_weight"` //人机类型权重
	RobotType       []int         `json:"robot_type" bson:"robot_type"`               //人机类型
	EmojiConfig     []EmojiConfig `json:"emoji_config" bson:"emoji_config"`           //人机表情配置
}

func (e Emoji) FindEmojiConfig(id int) EmojiConfig {
	for _, v := range e.EmojiConfig {
		if v.Id == id {
			return v
		}
	}
	return EmojiConfig{}
}

// 人机表情配置
type EmojiConfig struct {
	Id int `json:"id" bson:"id"` //人机类型
	//*******************************
	DeskSendProb   int   `json:"desk_send_prob" bson:"desk_send_prob"`     //上桌发送概率
	DeskSendWeight []int `json:"desk_send_weight" bson:"desk_send_weight"` //表情权重
	DeskSendDelay  []int `json:"desk_send_delay" bson:"desk_send_delay"`   //触发延迟
	//*******************************
	GameStartProb   int   `json:"game_start_prob" bson:"game_start_prob"`     //游戏开始发送概率
	GameStartWeight []int `json:"game_start_weight" bson:"game_start_weight"` //表情权重
	GameStartDelay  []int `json:"game_start_delay" bson:"game_start_delay"`   //触发延迟
	//*******************************
	WaitLongProb   int   `json:"wait_long_prob" bson:"wait_long_prob"`     //等待太久发送概率
	WaitLongWeight []int `json:"wait_long_weight" bson:"wait_long_weight"` //表情权重
	WaitLongDelay  []int `json:"wait_long_delay" bson:"wait_long_delay"`   //触发延迟
	//*******************************
	ReplyProb   int   `json:"reply_prob" bson:"reply_prob"`     //回复表情
	ReplyWeight []int `json:"reply_weight" bson:"reply_weight"` //表情权重
	ReplyDelay  []int `json:"reply_delay" bson:"reply_delay"`   //触发延迟
	// *******************************
	OtherSeeProb   int   `json:"other_see_prob" bson:"other_see_prob"`     //其他玩家看牌后
	OtherSeeWeight []int `json:"other_see_weight" bson:"other_see_weight"` //表情权重
	OtherSeeDelay  []int `json:"other_see_delay" bson:"other_see_delay"`   //触发延迟
	// *******************************
	OtherRaiseProb   int   `json:"other_raise_prob" bson:"other_raise_prob"`     //其他玩家加注
	OtherRaiseWeight []int `json:"other_raise_weight" bson:"other_raise_weight"` //表情权重
	OtherRaiseDelay  []int `json:"other_raise_delay" bson:"other_raise_delay"`   //触发延迟
	// *******************************
	OtherBiProb   []int `json:"other_bi_prob" bson:"other_bi_prob"`     //其他玩家比牌
	OtherBiWeight []int `json:"other_bi_weight" bson:"other_bi_weight"` //表情权重
	OtherBiDelay  []int `json:"other_bi_delay" bson:"other_bi_delay"`   //触发延迟
	// *******************************
	SelfBiWinProb   int   `json:"self_bi_win_prob" bson:"self_bi_win_prob"`     //自己比牌赢
	SelfBiWinWeight []int `json:"self_bi_win_weight" bson:"self_bi_win_weight"` //表情权重
	SelfBiWinDelay  []int `json:"self_bi_win_delay" bson:"self_bi_win_delay"`   //触发延迟
	// *******************************
	MoPreCardProb   int   `json:"mo_pre_card_prob" bson:"mo_pre_card_prob"`     //摸上家牌
	MoPreCardWeight []int `json:"mo_pre_card_weight" bson:"mo_pre_card_weight"` //表情权重
	MoPreCardDelay  []int `json:"mo_pre_card_delay" bson:"mo_pre_card_delay"`   //触发延迟
	// *******************************
	OtherChuCardProb   int   `json:"other_chu_card_prob" bson:"other_chu_card_prob"`     //其他玩家出牌
	OtherChuCardWeight []int `json:"other_chu_card_weight" bson:"other_chu_card_weight"` //表情权重
	OtherChuCardDelay  []int `json:"other_chu_card_delay" bson:"other_chu_card_delay"`   //触发延迟
}

func (ec *EmojiConfig) IsValid() bool {
	if ec.Id == 0 {
		return false
	} else {
		return true
	}
}

func (e *Emoji) Save() {
	Upsert(Emojis, bson.M{"_id": e.Id}, e)
}

func GetEmojiList() []Emoji {
	var list []Emoji
	ListByQ(Emojis, nil, &list)
	return list
}
