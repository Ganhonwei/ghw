package data

import (
	"gopkg.in/mgo.v2/bson"
)

type ShareData struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`            //名称
	LastLogin int64  `json:"lastLogin" bson:"last_login"` //上次登录
	Recharge  int64  `json:"recharge" bson:"recharge"`    //充值金额
	Receive   bool   `json:"receive" bson:"receive"`      //是否领取奖励
	Ctime     int64  `json:"ctime" bson:"ctime"`          //创建时间
}

type Share struct {
	Id            int32  `json:"id" bson:"_id"`                       // id
	Recharge      int32  `json:"recharge" bson:"recharge"`            // 充值领奖要求
	FirstRecharge int32  `json:"firstRecharge" bson:"first_recharge"` // 下线首单充值奖励
	MaxPeople     int32  `json:"maxPeople" bson:"max_people"`         // 下线人数上限
	Url           string `json:"url" bson:"url"`                      // 分享地址
	PCUrl         string `json:"pcUrl" bson:"pc_url"`                 // pc分享地址
	FirstReward   int32  `json:"firstReward" bson:"first_reward"`     // 第一级奖励
	SecondReward  int32  `json:"secondReward" bson:"second_reward"`   // 第二季奖励
}

type ShareAddr struct {
	Id   string `json:"id" bson:"_id"`    // id
	Link string `json:"link" bson:"link"` // 分享链接
}

// 打码量记录
type ShareAmountLog struct {
	Id        string        `json:"id" bson:"_id"`
	Ctime     int64         `json:"ctime" bson:"ctime"`          // 时间
	BetAmount int64         `json:"betAmount" bson:"bet_amount"` // 总打码量
	Detail    []ShareDetail `json:"detail" bson:"detail"`        // 详情
}

type ShareDetail struct {
	BetAmount int64  `json:"betAmount" bson:"bet_amount"` //总打码量
	UserId    string `json:"userId" bson:"user_id"`
	Name      string `json:"name" bson:"name"`     //名称
	Income    int64  `json:"income" bson:"income"` //今日进账
}

// 下级玩家打码量
type ShareBetAmount struct {
	Userid   string `json:"userid" bson:"_id"`        // 玩家id
	Name     string `json:"name" bson:"name"`         // 名字
	Superior string `json:"superior" bson:"superior"` // 上级玩家id
	Score    int64  `json:"score" bson:"score"`       // 打码量
	Ctime    int64  `json:"ctime" bson:"ctime"`       // 时间
}

func (s *Share) Save() {
	Upsert(Shares, bson.M{"_id": s.Id}, s)
}

func (s *ShareAddr) Save() {
	Upsert(ShareAddrs, bson.M{"_id": s.Id}, s)
}

// GetShareList 获取列表
func GetShareList() []Share {
	var list []Share
	// q := bson.M{}
	ListByQ(Shares, nil, &list)
	return list
}

// GetShareAddrList 获取列表
func GetShareAddrList() []ShareAddr {
	var list []ShareAddr
	// q := bson.M{}
	ListByQ(ShareAddrs, nil, &list)
	return list
}

func (s *ShareBetAmount) Save() {
	a := new(ShareBetAmount)
	Get(ShareAmounts, s.Userid, a)
	if a.Userid != "" {
		a.Ctime = s.Ctime
		a.Score += s.Score
	}
	if a.Userid != "" {
		Upsert(ShareAmounts, bson.M{"_id": s.Userid}, a)
	} else {
		Upsert(ShareAmounts, bson.M{"_id": s.Userid}, s)
	}
}

// 代理活动
type ShareAgent struct {
	TeamLv           int32   `json:"teamLv" bson:"team_lv"`                      // 团队等级
	TeamAgents       int32   `json:"teamAgents" bson:"team_agents"`              // 当前团队总人数
	TeamBets         int64   `json:"teamBets" bson:"team_bets"`                  // 当前团队打码量
	ShareUsers       int32   `json:"shareUsers" bson:"share_users"`              // 当前累计有效人头数
	ShareUsersPrizes []int32 `json:"shareUsersPrizes" bson:"share_users_prizes"` // 当前已达成累计有效人头数奖励
	ShareSuper       bool    `json:"shareSuper" bson:"share_super"`              // 是否已计算为有效人头

	// 奖励领取金额记录
	EarningsBetTotal      int64 `json:"earningsBetTotal" bson:"earnings_bet_total"`           // 已领取的打码返佣金额
	EarningsBetUnclaimed  int64 `json:"earningsBetUnclaimed" bson:"earnings_bet_unclaimed"`   // 还没领取的打码返佣金额
	EarningsBetPendding   int64 `json:"earningsBetPendding" bson:"earnings_bet_pendding"`     // 处理中的打码返佣金额(0点后加到unclaimed)(单位:厘)
	EarningsUserTotal     int64 `json:"earningsUserTotal" bson:"earnings_user_total"`         // 已领取的人头奖励和累计任务返佣金额
	EarningsUserUnclaimed int64 `json:"earningsUserUnclaimed" bson:"earnings_user_unclaimed"` // 还没领取的人头奖励和累计任务返佣金额
	EarningsUserPendding  int64 `json:"earningsUserPendding" bson:"earnings_user_pendding"`   // 处理中的人头奖励和累计任务返佣金额(0点后加到unclaimed)(单位:分)

	TodayEarningsUser   int32    `json:"todayEarningsUser" bson:"today_earnings_user"`      // 今日领取单个人头奖数 -> 单个人头奖的单日上限（人）
	TodayTakeItypeDates []string `json:"todayTakeItypeDates" bson:"today_take_itype_dates"` // 今日手动领取奖励类型[日期列表]记录
}
