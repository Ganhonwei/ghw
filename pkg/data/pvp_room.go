package data

import "github.com/globalsign/mgo/bson"

// 对战房基础配置
type PvpRoom struct {
	Id             int32   `bson:"_id" json:"id"`                           // id
	PvpChargeLimit int     `json:"pvpChargeLimit" bson:"pvp_charge_limit"`  // 开启功能充值要求
	GameStartWait  int     `json:"gameStartWait" bson:"game_start_wait"`    // 房间等待时间（秒）
	TpFunInitScore int     `json:"tpFunInitScore" bson:"tp_fun_init_score"` // TP娱乐模式初始分
	RmFunInitScore int     `json:"rmFunInitScore" bson:"rm_fun_init_score"` // RM娱乐模式初始分
	AbFunInitScore int     `json:"abFunInitScore" bson:"ab_fun_init_score"` // AB娱乐模式初始分
	TpGoldRound    []int   `json:"tpGoldRound" bson:"tp_gold_round"`        // TP真金模式局数
	RmGoldRound    []int   `json:"rmGoldRound" bson:"rm_gold_round"`        // RM真金模式局数
	AbGoldRound    []int   `json:"abGoldRound" bson:"ab_gold_round"`        // AB真金模式局数
	TpFunRoundCost [][]int `json:"tpFunRoundCost" bson:"tp_fun_round_cost"` // TP娱乐模式局数和费用
	RmFunRoundCost [][]int `json:"rmFunRoundCost" bson:"rm_fun_round_cost"` // RM娱乐模式局数和费用
	AbFunRoundCost [][]int `json:"abFunRoundCost" bson:"ab_fun_round_cost"` // AB娱乐模式局数和费用
}

// 保存对战房基础配置
func (t *PvpRoom) Save() bool {
	return Upsert(PvpRooms, bson.M{"_id": t.Id}, t)
}

// GetPvpRoomList 获取对战房基础配置列表
func GetPvpRoomList() []PvpRoom {
	var list []PvpRoom
	q := bson.M{}
	ListByQ(PvpRooms, q, &list)
	return list
}

// dbms 存储私人房间信息
type PrivRoom struct {
	Code   string
	Roomid string
	Gtype  int32
	Rtype  int32
}
