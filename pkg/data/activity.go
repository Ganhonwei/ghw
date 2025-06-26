package data

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"goserver/gen/pb"

	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

const (
	ActivityLoginPopKey = "activity_login_pop_%s_%d"
	ActivityLobbyPopKey = "activity_lobby_pop_%s_%d"
)

// Activity 活动配置
type Activity struct {
	Id        string    `bson:"_id" json:"id"`                //ID
	Type      int32     `bson:"type" json:"type"`             //类型0,1,2
	Del       int       `bson:"del" json:"del"`               //是否移除
	Title     string    `bson:"title" json:"title"`           //标题内容
	Content   string    `bson:"content" json:"content"`       //活动内容
	StartTime time.Time `bson:"start_time" json:"start_time"` //开始时间
	// EndTime   time.Time `bson:"end_time" json:"end_time"`     //结束时间
	Ctime time.Time `bson:"ctime" json:"ctime"` //创建时间
	// 首冲
	F_price     uint32 `bson:"f_price" json:"f_price"`       // 价格(分)
	F_number    uint32 `bson:"f_number" json:"f_number"`     // 首冲获得
	F_firstDay  uint32 `bson:"first_day" json:"first_day"`   // 第一天获得
	F_secondDay uint32 `bson:"second_day" json:"second_day"` // 第二天获得
	F_thirdDay  uint32 `bson:"third_day" json:"third_day"`   // 第三天获得
	F_forthDay  uint32 `bson:"forth_day" json:"forth_day"`   // 第四天获得
	F_fifthDay  uint32 `bson:"fifth_day" json:"fifth_day"`   // 第五天获得
	// 充值活动
	N_price      int32 `bson:"n_price" json:"n_price"`           // 价格
	N_number     int32 `bson:"n_number" json:"n_number"`         // 获得的crash
	N_proportion int32 `bson:"n_proportion" json:"n_proportion"` // 赠送比例
	// 周卡
	W_price      int32 `bson:"w_price" json:"w_price"`           // 价格
	W_number     int32 `bson:"w_number" json:"w_number"`         // 每日赠送的cash
	W_proportion int32 `bson:"w_proportion" json:"w_proportion"` // 赠送比例
	W_index      int32 `bson:"w_index" json:"w_index"`           // 排序
	W_duringDay  int32 `bson:"w_during_day" json:"w_during_day"` // 持续时间
	W_daily      int32 `bson:"w_daily" json:"w_daily"`           // 每日赠送bonus
}

// 签到
type DailySign struct {
	Id     int32 `bson:"_id" json:"id"`        // 第几天
	Ctype  int32 `bson:"ctype" json:"ctype"`   // 道具类型
	Number int32 `bson:"number" json:"number"` // 数量
	Price  int32 `json:"price" bson:"price"`   // 价格
}

// 金银铜卡
type WeekCard struct {
	Id              int32     `bson:"_id" json:"id"`                              // id
	ReceiveTimes    int32     `bson:"receive_times" json:"receive_times"`         // 已领取次数
	LastReceiveTime time.Time `bson:"last_receive_time" json:"last_receive_time"` // 上次领取时间
	// CanGet          int32     `bson:"can_get" json:"can_get"`                     // 可领取次数
	Get      bool  `json:"get" bson:"get"`            //是否可领取
	OverTime int64 `json:"overTime" bson:"over_time"` // 过期时间
}

// 周卡
type WeeklyCard struct {
	ID     int32   `json:"id" bson:"id"`         //id
	Price  int32   `json:"price" bson:"price"`   //价格
	Reward int32   `json:"reward" bson:"reward"` //奖励
	Give   []int32 `json:"give" bson:"give"`     //每日天赠送
}

// 在线奖励
type OnlineReward struct {
	Id         int32     `bson:"_id" json:"id"`                  // id
	OnlineTime int64     `bson:"online_time" json:"online_time"` // 在线时间
	Reward     [][]int32 `bson:"reward" json:"reward"`           // 奖励
}

// 限时礼包
type LimitedGift struct {
	Id         int32 `bson:"_id" json:"id"`                 // id
	Switch     int   `json:"switch" bson:"switch"`          // 开关
	Price      int32 `json:"price" bson:"price"`            // 价格
	Reward     int64 `json:"reward" bson:"reward"`          // 奖励
	DuringTime int64 `json:"duringTime" bson:"during_time"` // 持续时间
	Recharge   int32 `json:"recharge" bson:"recharge"`      // 充值要求
	GameRound  int32 `json:"gameRound" bson:"game_round"`   // 游戏局数要求
}

// 破产礼包
type BreakingGift struct {
	Gtype    int32  // 游戏类型
	GameId   string // 配置id
	BuyTimes int    // 当日购买次数
}

// 破产礼包充值
type BreakingGiftPay struct {
	Name  string //礼包名称
	Price uint32 //价格
	Cash  uint32 //金额
	Give  uint32 //赠送
}

// 大富翁
type ScratchTickets struct {
	CostBonus int64                       `json:"costBonus" bson:"cost_bonus"` // 累计消耗bonus
	Pokers    []uint32                    `json:"pokers" bson:"pokers"`        // 剩余扑克牌
	Version   int32                       `json:"version" bson:"version"`      // 版本
	Tickets   map[int]ScratchTicketDetail `json:"tickets" bson:"tickets"`      // 刮奖券
}

type ScratchTicketDetail struct {
	Id      int      `bson:"id" json:"id"`           // id
	Scratch []uint32 `json:"scratch" bson:"scratch"` // 刮开的奖
}

// 大富翁配置
type ScratchTicketConfig struct {
	Id         int                 `bson:"_id" json:"id"`                 // id 奖励等级
	NextId     int                 `json:"nextId" bson:"next_id"`         // 下一等级
	Cost       int64               `json:"cost" bson:"cost"`              // 消耗bonus
	Cell       int                 `json:"cell" bson:"cell"`              // 刮奖格子数
	Jackpot    []int64             `json:"jackpot" bson:"jackpot"`        // 奖金
	Scope      []int64             `json:"scope" bson:"scope"`            // 奖励范围
	Rule       []int               `json:"rule" bson:"rule"`              // 对应配置
	RuleConfig []ScratchTicketRule `json:"ruleConfig" bson:"rule_config"` // 中奖权重
}

type ScratchTicketRule struct {
	Id        int   `bson:"_id" json:"id"`          // id 奖励等级
	WeightPro []int `json:"weight1" bson:"weight1"` // 中奖概率
}

// 玩游戏分享数据
type PlayShareAct struct {
	Config          PlayAndDrawActivity `json:"config" bson:"config"`                     // 配置
	RealPlayTimes   int                 `json:"realPlayTimes" bson:"real_play_times"`     // 当前玩游戏次数
	RealInviteTimes int                 `json:"realInviteTimes" bson:"real_invite_times"` // 当前邀请次数
	PlayTimes       int                 `json:"playTimes" bson:"play_times"`              // 玩游戏次数
	InviteFriends   []string            `json:"inviteFriends" bson:"invite_friends"`      // 邀请玩家数(设备码)
	PlayDraws       int                 `json:"playDraws" bson:"play_draws"`              // 玩游戏剩余抽奖次数
	PlayUsed        int                 `json:"playUsed" bson:"play_used"`                // 使用的抽奖次数
	ShareDraws      int                 `json:"shareDraws" bson:"share_draws"`            // 分享抽奖剩余次数
	ShareUsed       int                 `json:"shareUsed" bson:"share_used"`              // 使用的分享抽奖次数
	PlayGet         int64               `json:"playGet" bson:"play_get"`                  // 玩游戏获得的奖励
	OverTime        int64               `json:"overTime" bson:"over_time"`                // 结束时间
	Rewards         int64               `json:"rewards" bson:"rewards"`                   // 获得的奖励
	Over            bool                `json:"over" bson:"over"`                         // 过期
}

// 玩游戏分享抽奖
type PlayAndDrawActivity struct {
	Id           int     `bson:"_id" json:"id"`                      // id
	ValidityTime int32   `json:"validityTime" bson:"validity_time"`  // 有效期(小时)
	MinDrawTimes int32   `json:"minDrawTimes" bson:"min_draw_times"` // 领到奖励的抽奖次数
	MaxPlayDraw  int32   `json:"MaxPlayDraw" bson:"max_play_draw"`   // 玩游戏抽奖上限
	MaxShareDraw int32   `json:"maxShareDraw" bson:"max_share_draw"` // 分享抽奖上限
	Reward       []int32 `json:"reward" bson:"reward"`               // 奖励
	RandSpan     float64 `json:"randSpan" bson:"rand_span"`          // 随机跨度
	JumpRate     float64 `json:"jumpRate" bson:"jump_rate"`          // 跃迁倍率
	DownRate     float64 `json:"downRate" bson:"down_rate"`          // 递减率
	PlayRounds   []int32 `json:"playRounds" bson:"play_rounds"`      // 游戏局数
	ShareFriends []int32 `json:"shareFriends" bson:"share_friends"`  // 分享朋友数
}

// 打码排行榜奖励
type ActivityBetRankPrize struct {
	Id                     string    `bson:"_id" json:"id"`
	Userid                 string    `bson:"userid" json:"userid"`
	Username               string    `bson:"username" json:"username"`
	VipLv                  int32     `bson:"vip_lv" json:"vipLv"`
	Photo                  string    `bson:"photo" json:"photo"`
	Robot                  bool      `bson:"robot" json:"robot"`
	RankType               int32     `bson:"rank_type" json:"rankType"`                               // 排行榜类型: 1日榜 2周榜 3月榜
	Rank                   int32     `bson:"rank" json:"rank"`                                        // 排名
	Bets                   int64     `bson:"bets" json:"bets"`                                        // 打码量
	Prize                  int64     `bson:"prize" json:"prize"`                                      // 奖池分成
	PrizeRate              string    `bson:"prize_rate" json:"prizeRate"`                             // 奖池分成比例
	Jackpot                int64     `bson:"jackpot" json:"jackpot"`                                  // 奖池
	JackpotPre             int64     `bson:"jackpot_pre" json:"jackpotPre"`                           // 奖池前置金额
	JackpotRepayPre        int64     `bson:"jackpot_repay_pre" json:"jackpotRepayPre"`                // 奖池已偿还前置金额
	JackpotSubsidyRepayPre int64     `bson:"jackpot_subsidy_repay_pre" json:"jackpotSubsidyRepayPre"` // 补贴已偿还奖池前置金额
	JackpotSubsidy         int64     `bson:"jackpot_subsidy" json:"jackpotSubsidy"`                   // 奖池补贴金额
	PrizeType              int32     `bson:"prize_type" json:"prizeType"`                             // 奖金类型: 1代表bonus,2代表cash,3代表withdrawable
	Stime                  int64     `bson:"stime" json:"stime"`                                      // 周期开始时间戳
	Etime                  int64     `bson:"etime" json:"etime"`                                      // 周期结束时间戳
	Ctime                  time.Time `bson:"ctime" json:"ctime"`                                      // 创建时间
	Received               int8      `bson:"received" json:"received"`                                // 是否领取 0否1是
	ReceiveTime            int64     `bson:"receive_time" json:"receiveTime"`                         // 领取时间
}

// Save 写入数据库
func (t *ActivityBetRankPrize) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncActivityBetRankPrize, []any{t})

	return Upsert(ActivityBetRankPrizes, bson.M{"_id": t.Id}, t)
}

// Get 查询获取记录
func (t *ActivityBetRankPrize) GetById() {
	GetByQ(ActivityBetRankPrizes, bson.M{"_id": t.Id}, t)
}

func ListActivityBetRankPrize(m bson.M) (prizes []ActivityBetRankPrize) {
	ListByQ(ActivityBetRankPrizes, m, &prizes)
	return
}

// 获取未领取打码排行榜奖励
func ListUserBetRankPrizeUnreceived(userid string) (prizes []ActivityBetRankPrize) {
	m := bson.M{"userid": userid, "received": 0}
	ListByQ(ActivityBetRankPrizes, m, &prizes)
	return
}

// 转盘活动每轮时间记录
type ActivityTurnTime struct {
	TurnStime int64 `bson:"_id" json:"turnStime"`        // 本轮转盘开始时间
	TurnEtime int64 `bson:"turn_etime" json:"turnEtime"` // 本轮转盘结束时间
}

// Save 写入数据库
func (t *ActivityTurnTime) Save() bool {
	return Insert(ActivityTurnTimes, t)
}

// 转盘活动抽奖记录
type ActivityTurnDrawLog struct {
	Id          string `bson:"_id" json:"id"`                   // 抽奖记录id
	TurnStime   int64  `bson:"turn_stime" json:"turnStime"`     // 本轮转盘开始时间
	TurnEtime   int64  `bson:"turn_etime" json:"turnEtime"`     // 本轮转盘结束时间
	Userid      string `bson:"userid" json:"userid"`            // 用户id
	Username    string `bson:"username" json:"username"`        // 用户名
	Photo       string `bson:"photo" json:"photo"`              // 头像
	ChannelId   string `bson:"channel_id" json:"channelId"`     // 用户渠道
	RegistArea  int8   `bson:"regist_area" json:"registArea"`   // 用户ab测试 0:A 1:B 2:C
	Rtime       int64  `bson:"rtime" json:"rtime"`              // 用户注册时间(时间戳秒)
	Reason      int32  `bson:"reason" json:"reason"`            // 0序幕礼盒,1免费次数抽中,2邀请次数抽中
	DrawScore   int64  `bson:"draw_score" json:"drawScore"`     // 抽中分数
	Ctime       int64  `bson:"ctime" json:"ctime"`              // 抽奖时间(时间戳秒)
	ScoreBefore int64  `bson:"score_before" json:"scoreBefore"` // 抽中前分数
	Score       int64  `bson:"score" json:"score"`              // 抽中后分数
	ScoreTarget int64  `bson:"score_target" json:"scoreTarget"` // 目标分数
	PrizeType   int32  `bson:"prize_type" json:"prizeType"`     // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	LuckyDraw   int32  `bson:"lucky_draw" json:"luckyDraw"`     // 1.触发有效邀请直接抽满概率 2.触发有效邀请直接抽满概率且直接抽满
}

// Save 写入数据库
func (t *ActivityTurnDrawLog) Save() bool {
	DataProducers.PublishDatas(mq.TopicSyncActivityTurnDrawLog, []any{t})

	return Insert(ActivityTurnDrawLogs, t)
}

// 转盘活动领奖记录
type ActivityTurnPrizeLog struct {
	Id                string `bson:"_id" json:"id"`                                // 抽奖记录id
	TurnStime         int64  `bson:"turn_stime" json:"turnStime"`                  // 本轮转盘开始时间
	TurnEtime         int64  `bson:"turn_etime" json:"turnEtime"`                  // 本轮转盘结束时间
	Userid            string `bson:"userid" json:"userid"`                         // 用户id
	ChannelId         string `bson:"channel_id" json:"channelId"`                  // 用户渠道
	RegistArea        int8   `bson:"regist_area" json:"registArea"`                // 用户ab测试 0:A 1:B 2:C
	Rtime             int64  `bson:"rtime" json:"rtime"`                           // 用户注册时间(时间戳秒)
	DrawedTimesFree   int32  `bson:"drawed_times_free" json:"drawedTimesFree"`     // 已使用免费抽奖次数
	DrawedTimesInvite int32  `bson:"drawed_times_invite" json:"drawedTimesInvite"` // 已使用邀请抽奖次数
	Gives             int64  `bson:"gives" json:"gives"`                           // 赠送序幕金
	Prize             int64  `bson:"prize" json:"prize"`                           // 奖金金额
	PrizeType         int32  `bson:"prize_type" json:"prizeType"`                  // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	Ctime             int64  `bson:"ctime" json:"ctime"`                           // 申请领奖时间(时间戳秒)
	State             int32  `bson:"state" json:"state"`                           // 审核状态: 0审核中,1通过,2拒绝
	Stime             int64  `bson:"stime" json:"stime"`                           // 审核时间(时间戳秒)
	Stype             int32  `bson:"stype" json:"stype"`                           // 审核类型:1机审,2人工审核
	Reason            string `bson:"reason" json:"reason"`                         // 机审结果原因
	AuditUser         string `bson:"audit_user" json:"auditUser"`                  // 审核人
	Remark            string `bson:"remark" json:"remark"`                         // 后台备注
}

// Save 写入数据库
func (t *ActivityTurnPrizeLog) Save() bool {
	DataProducers.PublishDatas(mq.TopicSyncActivityTurnPrizeLog, []any{t})

	return Upsert(ActivityTurnPrizeLogs, bson.M{"_id": t.Id}, t)
}

// GetById 根据id查询
func (t *ActivityTurnPrizeLog) GetById() {
	GetByQ(ActivityTurnPrizeLogs, bson.M{"_id": t.Id}, t)
}

// 根据id查询转盘奖励列表
func ListActivityTurnPrizeLogByIds(ids []string) (prizes []*ActivityTurnPrizeLog) {
	m := bson.M{"_id": bson.M{"$in": ids}}
	ListByQ(ActivityTurnPrizeLogs, m, &prizes)
	return
}

// 审核操作
type ActivityTurnPrizeOperate struct {
	PrizeIds []string `json:"prizeIds"` // 订单ID
	Op       int      `json:"op"`       // 操作 1通过2拒绝
	UserName string   `json:"username"` // 操作人
}

// Save 写入数据库
func (t *PlayAndDrawActivity) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(PlayAndDrawActivitys, bson.M{"_id": t.Id}, t)
}

func GetPlayShareList() []PlayAndDrawActivity {
	//t.Ctime = bson.Now()
	var list []PlayAndDrawActivity
	ListByQ(PlayAndDrawActivitys, nil, &list)
	return list
}

// Save 写入数据库
func (t *ScratchTicketConfig) Save() bool {
	//t.Ctime = bson.Now()
	return Upsert(ScratchTicketss, bson.M{"_id": t.Id}, t)
}

func GetScratchTicketConfigList() []*ScratchTicketConfig {
	var list []*ScratchTicketConfig
	ListByQ(ScratchTicketss, nil, &list)
	return list
}

// 保存签到数据 TODO 后面会该为配置表
func (t *DailySign) Save() bool {
	return Upsert(DailySigns, bson.M{"_id": t.Id}, t)
}

// 保存在线奖励数据 TODO 后面会该为配置表
func (t *OnlineReward) Save() bool {
	return Upsert(OnlineRewards, bson.M{"_id": t.Id}, t)
}

// 限时礼包
func (t *LimitedGift) Save() bool {
	return Upsert(LimitedGifts, bson.M{"_id": t.Id}, t)
}

// Save 保存消息记录
func (t *Activity) Save() bool {
	//t.Id = ObjectIdString(bson.NewObjectId())
	//t.Ctime = bson.Now()
	return Insert(Activitys, t)
}

// GetActivityList 获取列表
func GetActivityList() []Activity {
	var list []Activity
	// q := bson.M{"del": 0, "end_time": bson.M{"$gt": bson.Now()}}
	q := bson.M{"del": 0}
	ListByQ(Activitys, q, &list)
	return list
}

// GetDailySignList 获取列表
func GetDailySignList() []DailySign {
	var list []DailySign
	// q := bson.M{}
	ListByQ(DailySigns, nil, &list)
	return list
}

// GetOnlineRewardList 获取列表
func GetOnlineRewardList() []OnlineReward {
	var list []OnlineReward
	// q := bson.M{}
	ListByQ(OnlineRewards, nil, &list)
	return list
}

// GetLimitedGiftList 获取列表
func GetLimitedGiftList() []LimitedGift {
	var list []LimitedGift
	// q := bson.M{}
	ListByQ(LimitedGifts, nil, &list)
	return list
}

func (w *WeeklyCard) Save() {
	Upsert(WeeklyCards, bson.M{"_id": w.ID}, w)
}

func GetWeeklyCardList() []WeeklyCard {
	var list []WeeklyCard
	ListByQ(WeeklyCards, nil, &list)
	return list
}

// LogActivity 玩家参与活动记录
type LogActivity struct {
	//Id      string    `bson:"_id" json:"id"`          //ID
	Userid string    `bson:"userid" json:"userid"` //玩家
	Actid  string    `bson:"actid" json:"actid"`   //activity id
	Type   int32     `bson:"type" json:"type"`     //类型0,1,2
	Prize  int64     `bson:"prize" json:"prize"`   //奖励数量
	Num    uint32    `bson:"num" json:"num"`       //完成次数
	Etime  time.Time `bson:"etime" json:"etime"`   //过期时间
	Jtime  time.Time `bson:"jtime" json:"jtime"`   //参与时间
	Utime  time.Time `bson:"utime" json:"utime"`   //update Time
	Ctime  time.Time `bson:"ctime" json:"ctime"`   //创建时间
}

// Save 保存消息记录
func (t *LogActivity) Save() bool {
	//t.Id = bson.NewObjectId().String()
	//t.Id = ObjectIdString(bson.NewObjectId())
	t.Ctime = bson.Now()
	//t.Etime = bson.Now().AddDate(0, 0, 7)
	return Insert(LogActivitys, t)
}

// Get 查询获取记录
func (t *LogActivity) Get() {
	GetByQ(LogActivitys, bson.M{"userid": t.Userid, "actid": t.Actid}, t)
}

// Has 记录是否存在
func (t *LogActivity) Has() bool {
	return Has(LogActivitys, bson.M{"userid": t.Userid, "actid": t.Actid})
}

// Update 更新记录
func (t *LogActivity) Update() bool {
	t.Utime = bson.Now()
	return Update(LogActivitys, bson.M{"userid": t.Userid, "actid": t.Actid},
		bson.M{"$set": bson.M{"utime": t.Utime}, "$inc": bson.M{"prize": t.Prize, "num": t.Num}})
}

// GetLogActivitys 获取玩家记录
func GetLogActivitys(userid string, page int, ids []string) ([]*LogActivity, error) {
	pageSize := 30
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	var list = make([]*LogActivity, 0)
	q := bson.M{"userid": userid}
	if len(ids) != 0 {
		q["actid"] = bson.M{"$in": ids}
	}
	err := LogActivitys.
		Find(q).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("none record")
	}
	return list, nil
}

// getLogActivityList 获取参加活动玩家记录
func getLogActivityList(page int, q bson.M) ([]*LogActivity, error) {
	pageSize := 30
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	var list = make([]*LogActivity, 0)
	err := LogActivitys.
		Find(q).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("none record")
	}
	return list, nil
}

// GetJoinActivityList 参加代理活动玩家列表
// func GetJoinActivityList(arg *pb.AgentActivity, act Activity) ([]*LogActivity, error) {
// 	q := bson.M{"actid": arg.GetActid()}
// 	switch act.Type {
// 	case int32(pb.ACT_TYPE0):
// 		q["prize"] = bson.M{"$lt": 250000}
// 	case int32(pb.ACT_TYPE1):
// 		q["num"] = bson.M{"$eq": 0}
// 	case int32(pb.ACT_TYPE2):
// 		q["num"] = bson.M{"$lt": 15}
// 	default:
// 		return nil, errors.New("type error")
// 	}
// 	return getLogActivityList(int(arg.GetPage()), q)
// }

// StatActivity 统计数据
// func StatActivity(actLog *LogActivity, act Activity) (int64, error) {
// 	switch act.Type {
// 	case int32(pb.ACT_TYPE0):
// 		endTime := utils.TimestampTodayTime()
// 		startTime := actLog.Jtime                     //参加开始时间
// 		if endTime.Sub(actLog.Jtime).Hours() > 7*24 { //参加超过7天,统计7日内
// 			startTime = endTime.AddDate(0, 0, -7)
// 		}
// 		q := bson.M{"agentid": actLog.Userid}
// 		q["day"] = bson.M{"$gte": utils.Time2DayDate(startTime),
// 			"$lt": utils.Time2DayDate(endTime)}
// 		glog.Debugf("userid %s, Type %d, q %#v", actLog.Userid, act.Type, q)
// 		return getAgentDayProfitCount2(q)
// 	case int32(pb.ACT_TYPE1):
// 		endTime := utils.TimestampTodayTime()
// 		startTime := endTime.AddDate(0, 0, -1)
// 		q := bson.M{"agentid": actLog.Userid}
// 		q["day"] = bson.M{"$eq": utils.Time2DayDate(startTime)}
// 		glog.Debugf("userid %s, Type %d, q %#v", actLog.Userid, act.Type, q)
// 		return getAgentDayProfitCount2(q)
// 	case int32(pb.ACT_TYPE2):
// 		ids, err := getAgentChilds(actLog.Userid)
// 		glog.Debugf("userid %s, ids %v, err %v", actLog.Userid, ids, err)
// 		if err != nil || len(ids) == 0 {
// 			return 0, err
// 		}
// 		endTime := utils.TimestampTodayTime()
// 		startTime := actLog.Jtime                      //参加开始时间
// 		if endTime.Sub(actLog.Jtime).Hours() > 30*24 { //参加超过30天,统计30日内
// 			startTime = endTime.AddDate(0, 0, -30)
// 		}
// 		q := bson.M{"agentid": bson.M{"$in": ids}}
// 		q["day"] = bson.M{"$gte": utils.Time2DayDate(startTime),
// 			"$lt": utils.Time2DayDate(endTime)}
// 		glog.Debugf("userid %s, Type %d, q %#v", actLog.Userid, act.Type, q)
// 		return getAgentChildsCount(q)
// 	default:
// 		return 0, errors.New("type error")
// 	}
// 	return 0, nil
// }

// 下级代理列表
func getAgentChilds(userid string) ([]string, error) {
	var ids []string
	if userid == "" {
		return ids, errors.New("userid error")
	}
	var list []bson.M
	selector := make(bson.M, 1)
	selector["_id"] = true
	q := bson.M{"agent": userid, "agent_state": bson.M{"$eq": 1}}
	err := PlayerUsers.
		Find(q).Select(selector).
		All(&list)
	if err != nil {
		return ids, err
	}
	if len(list) == 0 {
		return ids, errors.New("none record")
	}
	for _, v := range list {
		if val, ok := v["_id"]; ok {
			ids = append(ids, val.(string))
		}
	}
	return ids, nil
}

// 代理收益统计
// func getAgentChildsCount(q bson.M) (int64, error) {
// 	list, err := agentDayProfitGroup2(q)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(list) == 0 {
// 		return 0, errors.New("none record")
// 	}
// 	var n int64
// 	for _, v := range list {
// 		var num int64
// 		if val, ok := v["profit"]; ok {
// 			num += utils.Int64(val)
// 		}
// 		if val, ok := v["profit_first"]; ok {
// 			num += utils.Int64(val)
// 		}
// 		if val, ok := v["profit_second"]; ok {
// 			num += utils.Int64(val)
// 		}
// 		if num >= 30000000 {
// 			n++
// 		}
// 	}
// 	return n, nil
// }

// GetLogActivityList 参加代理活动玩家列表
// func GetLogActivityList(arg *pb.AgentActivityNotice) ([]*LogActivity, error) {
// 	q := bson.M{"actid": arg.GetActid(), "type": arg.GetType()}
// 	return getLogActivityList(int(arg.GetPage()), q)
// }

// 判断能不能买
func (a *Activity) CanOrder(user *User, id string, ctype int32) bool {
	switch a.Type {
	case int32(pb.ACT_TYPE0):
		return user.FirstRecharge == ""
	case int32(pb.ACT_TYPE1):
		return user.NoviceGift == ""
	case int32(pb.ACT_TYPE2):
		week, err := strconv.Atoi(id)
		if err != nil {
			return false
		}
		return user.WeeklyCardMap[int32(week)] == nil
	}
	return false
}

// 充值之后
func (act *Activity) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.ACTIVITY
	switch ctype {
	case RECHARGE_GIFT:
		user.NoviceGift = id
		if save {
			user.UpdateNovice()
		}
		rsp.Atype = pb.ACT_TYPE1
	case METAL_CARD:
		// weekid, _ := strconv.Atoi(id)
		// cardData := &WeekCard{
		// 	Id:              act.Id,
		// 	LastReceiveTime: utils.LocalTime(),
		// 	OverTime:        utils.LocalTime().Unix() + int64(7*24*time.Hour.Seconds()),
		// 	CanGet:          1,
		// 	ReceiveTimes:    1,
		// }
		// if user.WeeklyCardMap == nil {
		// 	user.WeeklyCardMap = make(map[int32]*WeekCard)
		// }
		// user.WeeklyCardMap[int32(weekid)] = cardData
		// if save {
		// 	user.UpdateWeekCard()
		// }
		// rsp.Atype = pb.ACT_TYPE2
	case FIRST_RECHARGE:
		user.FirstRecharge = act.Id
		user.Receive = true
		user.ReceiveDay = 1
		if save {
			user.UpdateFirstRecharge()
		}
		rsp.Atype = pb.ACT_TYPE0
	default:
		glog.Errorf("user %s recharge no fount acttype, actId:%s", user.Userid, id)
		rsp.Error = pb.NotFoundActType
		return rsp, fmt.Errorf("no fount acttype")
	}
	glog.Infof("user %s recharge success, actId:%s", user.Userid, id)
	return rsp, nil
}

// 完善订单
func (act *Activity) BuildPayOrder(user *User, order *PayOrder, game bool) {
	switch act.Type {
	case int32(pb.ACT_TYPE0):
		order.Amount = uint32(act.F_price)
		order.Score = uint32(act.F_number)
		order.OtherPresent = utils.String(act.F_firstDay)
		order.ShopType = FIRST_RECHARGE
		order.ShopId = act.Id
	case int32(pb.ACT_TYPE1):
		order.Amount = uint32(act.N_price)
		order.Score = uint32(act.N_number)
		order.OtherPresent = utils.String(int64(act.N_number) * int64(act.N_proportion) / 100)
		order.ShopType = RECHARGE_GIFT
		order.ShopId = act.Id
	case int32(pb.ACT_TYPE2):
		order.Amount = uint32(act.W_price)
		order.Score = uint32(act.W_number)
		order.OtherPresent = utils.String(act.W_daily)
		order.ShopType = METAL_CARD
		order.ShopId = act.Id
	}
	order.ShopName = act.Title
	if game {
		order.ShopName = act.Title + "(局内充值)"
	}
}

func (act *Activity) RechageLType() int32 {
	switch act.Type {
	case int32(pb.ACT_TYPE0):
		return int32(pb.LOG_TYPE57)
	case int32(pb.ACT_TYPE1):
		return int32(pb.LOG_TYPE58)
	case int32(pb.ACT_TYPE2):
		return int32(pb.LOG_TYPE59)
	}
	return 0
}

// 判断能不能买
func (a *WeeklyCard) CanOrder(user *User, id string, ctype int32) bool {
	if v, ok := user.WeeklyCardMap[a.ID]; ok {
		return v.OverTime == -1
	}
	return user.WeeklyCardMap[a.ID] == nil
}

// 充值之后
func (act *WeeklyCard) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.WEEKCARD
	t := utils.LocalTime().Unix() + int64(7*24*time.Hour.Seconds())
	ovt := time.Unix(t, 0) // 过期时间
	cardData := &WeekCard{
		Id:              act.ID,
		LastReceiveTime: utils.LocalTime(),
		OverTime:        utils.TimestampAppiontZero(ovt, time.Local),
		// CanGet:          1,
		Get:          false,
		ReceiveTimes: 1,
	}
	if user.WeeklyCardMap == nil {
		user.WeeklyCardMap = make(map[int32]*WeekCard)
	}
	user.WeeklyCardMap[act.ID] = cardData
	if save {
		user.UpdateWeekCard()
	}
	rsp.Atype = pb.ACT_TYPE2
	glog.Infof("user %s recharge success, actId:%s", user.Userid, id)
	return rsp, nil
}

// 完善订单
func (act *WeeklyCard) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.Amount = uint32(act.Price)
	order.Score = uint32(act.Reward)
	order.ShopType = METAL_CARD
	order.ShopId = fmt.Sprintf("%d", act.ID)
	order.ShopName = fmt.Sprintf("周卡%d", act.ID)
	wid := fmt.Sprintf("lbzk%d", act.ID)
	card := table.GetTables().GiftRechargeTable.Get(wid)
	// 赠送
	if len(card.DailyGive) > user.RegistArea {
		gives := card.DailyGive[user.RegistArea]
		// 赠送bonus
		order.OtherPresent = utils.String(int64(act.Reward) * int64(gives.Nums[0]) / 10000)
		// 赠送cash
		order.GiveCash = int64(act.Reward) * int64(gives.Nums[1]) / 10000
		// 赠送withdrawable
		order.GiveWithdrawal = int64(act.Reward) * int64(gives.Nums[2]) / 10000
	}
}

func (act *WeeklyCard) RechageLType() int32 {
	return int32(pb.LOG_TYPE59)
}

// --------------------------限时礼包--------------------------------
// 判断能不能买
func (a *LimitedGift) CanOrder(user *User, id string, ctype int32) bool {
	return user.LimitedGiftId == a.Id && user.LimitedGiftOverTime >= utils.LocalTime().Local().UnixMilli()
}

// 充值之后
func (act *LimitedGift) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.LIMITEDGIFT

	user.LimitedGiftId = 0
	user.LimitedGiftOverTime = 0
	user.OverlimitedGift = append(user.OverlimitedGift, act.Id)

	if save {
		user.UpdateLimitedGift()
	}
	rsp.Atype = pb.ACT_TYPE3
	glog.Infof("user %s recharge success, limitId:%s", user.Userid, id)
	return rsp, nil
}

// 完善订单
func (act *LimitedGift) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.Amount = uint32(act.Price)
	order.Score = uint32(act.Reward)
	order.ShopType = LIMITED_GIFT
	order.ShopId = fmt.Sprintf("%d", act.Id)
	order.ShopName = fmt.Sprintf("限时礼包%d", act.Id)
}

func (act *LimitedGift) RechageLType() int32 {
	return int32(pb.LOG_TYPE84)
}

// ------------------------ 破产礼包 ----------------------------------
// 判断能不能买
func (a *BreakingGiftPay) CanOrder(user *User, id string, ctype int32) bool {
	return true
}

// 充值之后
func (act *BreakingGiftPay) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.BREAKING

	exist := false
	for _, g := range user.BreakingGift {
		if g.Gtype == user.BreakGtype && g.GameId == user.BreakGameId {
			exist = true
			g.BuyTimes++
			break
		}
	}

	if !exist {
		b := &BreakingGift{Gtype: user.BreakGtype, GameId: user.BreakGameId, BuyTimes: 1}
		user.BreakingGift = append(user.BreakingGift, b)
	}

	if save {
		user.UpdateBreakingGift()
	}
	rsp.Atype = pb.ACT_TYPE4
	glog.Infof("user %s recharge breaking gift success", user.Userid)
	return rsp, nil
}

// 完善订单
func (act *BreakingGiftPay) BuildPayOrder(user *User, order *PayOrder, game bool) {
	order.Amount = uint32(act.Price)
	order.Score = uint32(act.Cash)
	//
	order.ShopType = BREAKING_GITF
	order.ShopName = act.Name

	order.OtherPresent = utils.String(act.Give)
	// if user.VBBank >= int64(act.Give) {
	// }
}

func (act *BreakingGiftPay) RechageLType() int32 {
	return int32(pb.LOG_TYPE86)
}

// ------------------------ 签到 ----------------------------------
// 判断能不能买
func (a *DailySign) CanOrder(user *User, id string, ctype int32) bool {
	day := a.Id
	if user.LoginTimes < int32(day) {
		return false
	}
	if user.SignDay&(1<<day) != 0 {
		return false
	}
	return a.Price > 0
}

// 充值之后
func (act *DailySign) RechargeAfter(user *User, id string, ctype int32, save bool) (*pb.PaySuccessNtf, error) {
	rsp := new(pb.PaySuccessNtf)
	rsp.Id = id
	rsp.Rtype = pb.SIGN

	user.SignDay |= (1 << act.Id)

	if save {
		user.UpdateSign()
	}
	// rsp.Atype = pb.ACT_TYPE4
	glog.Infof("user %s recharge sign success, day:%d", user.Userid, act.Id)
	return rsp, nil
}

// 完善订单
func (act *DailySign) BuildPayOrder(user *User, order *PayOrder, game bool) {
	give := act.Number - act.Price
	if give <= 0 {
		give = 0
	}
	order.Amount = uint32(act.Price)
	order.Score = uint32(act.Price)
	order.ShopType = SIGN_RECHARGE
	order.ShopId = utils.String(act.Id)
	order.ShopName = fmt.Sprintf("第%d天充值", act.Id)

	order.OtherPresent = utils.String(give)
	// if user.VBBank >= int64(give) {
	// }
}

func (act *DailySign) RechageLType() int32 {
	return int32(pb.LOG_TYPE129)
}

// 代理 打码/人头 奖励记录表 BettingCommission/ReferralBonus
type ShareAgentIncomeRecord struct {
	Id         string `bson:"_id" json:"id"`                 // id
	Userid     string `bson:"userid" json:"userid"`          // 用户id
	SuperId    string `bson:"super_id" json:"superId"`       // 上级id
	Lv         int32  `bson:"lv" json:"lv"`                  // 属于几级代理
	TeamLv     int32  `bson:"team_lv" json:"team_lv"`        // 获得奖励的上级团队等级
	Itype      int32  `bson:"itype" json:"itype"`            // 1.打码奖励,2.人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖励
	Amount     int64  `bson:"amount" json:"amount"`          // 返佣奖励金额(毫:1分=10厘=100毫)
	AmountType int32  `bson:"amount_type" json:"amountType"` // 返佣奖励类型: 1:bonus,2:cash,3:withdrawable
	WaterId    string `bson:"water_id" json:"waterId"`       // 打码对局id/人头订单id
	Gtype      int32  `bson:"gtype" json:"gtype"`            // 打码奖励游戏类型
	Bets       int64  `bson:"bets" json:"bets"`              // 打码奖励打码量
	Idate      string `bson:"idate" json:"idate"`            // 日期字符串 2025-01-01
	Ctime      int64  `bson:"ctime" json:"ctime"`            // 创建时间戳毫秒
}

// Save 写入数据库
func (t *ShareAgentIncomeRecord) Save() {
	DataProducers.PublishDatas(mq.TopicSyncShareAgentIncomeRecord, []any{t})
	Insert(ShareAgentIncomeRecords, t)
}

// 代理 打码/人头 奖励领取记录表
type ShareAgentIncomeRecordTackLog struct {
	Id         string `bson:"_id" json:"id"`                 // id
	SuperId    string `bson:"super_id" json:"superId"`       // 用户id
	Amount     int64  `bson:"amount" json:"amount"`          // 领取奖励金额(分)
	AmountType int32  `bson:"amount_type" json:"amountType"` // 返佣奖励类型: 1:bonus,2:cash,3:withdrawable
	Idate      string `bson:"idate" json:"idate"`            // 日期字符串 2025-01-01
	Itype      int32  `bson:"itype" json:"itype"`            // 1.打码奖励,2人数人头奖励,3.累计任务人头奖励,4.受邀者人头奖励
	Ctime      int64  `bson:"ctime" json:"ctime"`            // 创建时间戳毫秒
}

// Save 写入数据库
func (t *ShareAgentIncomeRecordTackLog) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncShareAgentIncomeRecordTackLog, []any{t})
	return Upsert(ShareAgentIncomeRecordTackLogs, bson.M{"_id": t.Id}, t)
}

// 礼包码
type GiftPackCode struct {
	Code           string   `bson:"_id" json:"id"`                         // id,码号
	Status         int32    `bson:"status" json:"status"`                  // 0.关 1.开 2.已过期
	GiftType       int32    `bson:"gift_type" json:"giftType"`             // 1.随机码 2.等额码
	ScoreType      int32    `bson:"score_type" json:"scoreType"`           // 奖金类型: 1.bonus, 2.cash, 3.withdrawable
	Score          int64    `bson:"score" json:"score"`                    // 总金额
	Pieces         int32    `bson:"pieces" json:"pieces"`                  // 份数
	ScoreReceived  int64    `bson:"score_received" json:"scoreReceived"`   // 已领取份数
	PiecesReceived int32    `bson:"pieces_received" json:"piecesReceived"` // 已领取金额
	FinishTime     int64    `bson:"finish_time" json:"finishTime"`         // 领完时间
	PieceMin       int64    `bson:"piece_min" json:"pieceMin"`             // 单份最小金额
	PieceMax       int64    `bson:"piece_max" json:"pieceMax"`             // 单份最大金额
	TimeOpen       int64    `bson:"time_open" json:"timeOpen"`             // 有效期开始秒
	TimeClose      int64    `bson:"time_close" json:"timeClose"`           // 有效期结束秒
	Utypes         []int32  `bson:"utypes" json:"utypes"`                  // 标签满足一个即可: 0:A 1:B 2:C, 10.新手 11.平民 12.普R 13.小R 14.中R 15.大R 16.超大R
	Ctime          int64    `bson:"ctime" json:"ctime"`                    // 创建时间秒
	Cuser          string   `bson:"cuser" json:"cuser"`                    // 创建人
	Remark         string   `bson:"remark" json:"remark"`                  // 礼包码标签
	Userids        []string `bson:"userids" json:"userids"`                // 已领取用户列表
}

// Save 写入数据库
func (t *GiftPackCode) Save() bool {
	return Upsert(GiftPackCodes, bson.M{"_id": t.Code}, t)
}

// Save 写入数据库
func (t *GiftPackCode) Get() {
	GetByQ(GiftPackCodes, bson.M{"_id": t.Code}, t)
}

// 礼包码领取记录
type GiftPackCodeTackLog struct {
	Id          string  `bson:"_id" json:"id"`                   // id
	Code        string  `bson:"code" json:"code"`                // 码号
	Userid      string  `bson:"userid" json:"userid"`            // userid
	GiftType    int32   `bson:"gift_type" json:"giftType"`       // 1.随机码 2.等额码
	ScoreTack   int64   `bson:"score_tack" json:"scoreTack"`     // 领取金额
	Score       int64   `bson:"score" json:"score"`              // 总金额
	ScoreType   int32   `bson:"score_type" json:"scoreType"`     // 奖金类型: 1.bonus, 2.cash, 3.withdrawable
	ScoreBefore int64   `bson:"score_before" json:"scoreBefore"` // 领取前剩余奖金
	ScoreAfter  int64   `bson:"score_after" json:"scoreAfter"`   // 领取后剩余奖金
	Utypes      []int32 `bson:"utypes" json:"utypes"`            // 用户满足标签 大中小超ABC类用户
	Ctime       int64   `bson:"ctime" json:"ctime"`              // 创建时间秒
}

// Save 写入数据库
func (t *GiftPackCodeTackLog) Save() bool {
	return Insert(GiftPackCodeTackLogs, t)
}

// 波动返水领取记录
type VolatilitySubsidy struct {
	Id          string `bson:"_id" json:"id"`                   // id
	Userid      string `bson:"userid" json:"userid"`            // 用户id
	First       bool   `bson:"first" json:"first"`              // 是否首次领
	FirstTime   int64  `bson:"first_time" json:"firstTime"`     // 首次领取时间戳毫秒
	Sdate       string `bson:"sdate" json:"sdate"`              // 日期字符串 2025-01-01
	Subsidy     int64  `bson:"subsidy" json:"subsidy"`          // 补贴金额
	SubsidyType int32  `bson:"subsidy_type" json:"subsidyType"` // 补贴金额类型: 1:bonus,2:cash,3:withdrawable
	Ctime       int64  `bson:"ctime" json:"ctime"`              // 创建时间戳毫秒
	RegistArea  int32  `bson:"regist_area" json:"registArea"`   // abc类
	Pays        int64  `bson:"pays" json:"pays"`                // 领取时充值
	Bets        int64  `bson:"bets" json:"bets"`                // 领取时打码量
}

// Save 写入数据库
func (t *VolatilitySubsidy) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncVolatilitySubsidy, []any{t})
	return Insert(VolatilitySubsidys, t)
}

// 波动返水各类水池
type VolatilitySubsidyPool struct {
	Id         string `bson:"_id" json:"id"`                 // id: date-ptype-registArea
	Date       string `bson:"date" json:"date"`              // 日期
	Ptype      int32  `bson:"ptype" json:"ptype"`            // 1.保底水池,2.静态水池,3.动态水池,4.已补贴金额
	RegistArea int32  `bson:"regist_area" json:"registArea"` // abc类
	Amount     int64  `bson:"amount" json:"amount"`          // 水池金额
	Utime      int64  `bson:"utime" json:"utime"`            // 更新时间
}

// Save 写入数据库
func (t *VolatilitySubsidyPool) Save() bool {
	return Upsert(VolatilitySubsidyPools, bson.M{"_id": t.Id}, t)
}

func LogVolatilitySubsidyPool(log *pb.VolatilitySubsidyPool) {
	pool := &VolatilitySubsidyPool{
		Id:         log.Id,
		Date:       log.Date,
		Ptype:      log.Ptype,
		RegistArea: log.RegistArea,
		Amount:     log.Amount,
		Utime:      log.Utime,
	}
	pool.Save()
}

// 累充转盘解锁记录
type CumRechargeWheelUnlockLog struct {
	Id     string `bson:"_id" json:"id"`        // id
	Userid string `bson:"userid" json:"userid"` // 用户id
	Amount int64  `bson:"amount" json:"amount"` // 充值金额
	Times  int32  `bson:"times" json:"times"`   // 解锁次数
	Ctime  int64  `bson:"ctime" json:"ctime"`   // 创建时间
}

// Save 写入数据库
func (t *CumRechargeWheelUnlockLog) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncCumRechargeWheelUnlockLog, []any{t})
	return Insert(CumRechargeWheelUnlockLogs, t)
}

// 累充转盘抽奖记录
type CumRechargeWheelSpinLog struct {
	Id           string `bson:"_id" json:"id"`                     // id
	Userid       string `bson:"userid" json:"userid"`              // 用户id
	Rewardid     int32  `bson:"rewardid" json:"rewardid"`          // 奖项id
	WheelType    int32  `bson:"wheel_type" json:"wheelType"`       // 转盘类型
	RewardType   int32  `bson:"reward_type" json:"rewardType"`     // 奖项类型
	RewardText   string `bson:"reward_text" json:"rewardText"`     // 奖项文本
	RewardAmount int32  `bson:"reward_amount" json:"rewardAmount"` // 奖品数量
	RewardValue  int64  `bson:"reward_value" json:"rewardValue"`   // 奖品价值
	DistType     int32  `bson:"dist_type" json:"distType"`         // 发放类型: 1代表bonus，2代表cash，3代表withdrawable
	Ctime        int64  `bson:"ctime" json:"ctime"`                // 创建时间
}

// Save 写入数据库
func (t *CumRechargeWheelSpinLog) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncCumRechargeWheelSpinLog, []any{t})
	return Insert(CumRechargeWheelSpinLogs, t)
}
