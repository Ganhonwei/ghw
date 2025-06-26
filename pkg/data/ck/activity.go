package ck

import (
	"fmt"
	"goserver/pkg/data"
	"time"

	"github.com/bytedance/sonic"
)

// 交易订单
type ActivityBetRankPrize struct {
	Ver                    int64     `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id                     string    `gorm:"column:id" json:"id"`
	Userid                 string    `gorm:"column:userid" json:"userid"`
	Username               string    `gorm:"column:username" json:"username"`
	VipLv                  int32     `gorm:"column:vip_lv" json:"vipLv"`
	Photo                  string    `gorm:"column:photo" json:"photo"`
	Robot                  bool      `gorm:"column:robot" json:"robot"`
	RankType               int32     `gorm:"column:rank_type" json:"rankType"`                               // 排行榜类型: 1日榜 2周榜 3月榜
	Rank                   int32     `gorm:"column:rank" json:"rank"`                                        // 排名
	Bets                   int64     `gorm:"column:bets" json:"bets"`                                        // 打码量
	Prize                  int64     `gorm:"column:prize" json:"prize"`                                      // 奖池分成
	PrizeRate              string    `gorm:"column:prize_rate" json:"prizeRate"`                             // 奖池分成比例
	Jackpot                int64     `gorm:"column:jackpot" json:"jackpot"`                                  // 奖池
	JackpotPre             int64     `gorm:"column:jackpot_pre" json:"jackpotPre"`                           // 奖池前置金额
	JackpotRepayPre        int64     `gorm:"column:jackpot_repay_pre" json:"jackpotRepayPre"`                // 奖池已偿还前置金额
	JackpotSubsidyRepayPre int64     `gorm:"column:jackpot_subsidy_repay_pre" json:"jackpotSubsidyRepayPre"` // 补贴已偿还奖池前置金额
	JackpotSubsidy         int64     `gorm:"column:jackpot_subsidy" json:"jackpotSubsidy"`                   // 奖池补贴金额
	PrizeType              int32     `gorm:"column:prize_type" json:"prizeType"`                             // 奖金类型: 1代表bonus,2代表cash,3代表withdrawable
	Stime                  int64     `gorm:"column:stime" json:"stime"`                                      // 周期开始时间戳
	Etime                  int64     `gorm:"column:etime" json:"etime"`                                      // 周期结束时间戳
	Ctime                  time.Time `gorm:"column:ctime" json:"ctime"`                                      // 创建时间
	Received               int8      `gorm:"column:received" json:"received"`                                // 是否领取 0否1是
	ReceiveTime            int64     `gorm:"column:receive_time" json:"receiveTime"`                         // 领取时间
}

func (*ActivityBetRankPrize) TableName() string {
	return "col_activity_betrank_prize"
}

func (*ActivityBetRankPrize) New() CkEntity {
	return new(ActivityBetRankPrize)
}

func (c *ActivityBetRankPrize) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.ActivityBetRankPrize)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 转盘活动抽奖记录
type ActivityTurnDrawLog struct {
	Ver         int64  `gorm:"column:ver" json:"ver"`                  // 插入时间戳
	Id          string `gorm:"column:id" json:"id"`                    // 抽奖记录id
	TurnStime   int64  `gorm:"column:turn_stime" json:"turnStime"`     // 本轮转盘开始时间
	TurnEtime   int64  `gorm:"column:turn_etime" json:"turnEtime"`     // 本轮转盘结束时间
	Userid      string `gorm:"column:userid" json:"userid"`            // 用户id
	Username    string `gorm:"column:username" json:"username"`        // 用户名
	Photo       string `gorm:"column:photo" json:"photo"`              // 头像
	ChannelId   string `gorm:"column:channel_id" json:"channelId"`     // 用户渠道
	RegistArea  int8   `gorm:"column:regist_area" json:"registArea"`   // 用户ab测试 0:A 1:B 2:C
	Rtime       int64  `gorm:"column:rtime" json:"rtime"`              // 用户注册时间(时间戳秒)
	Reason      int32  `gorm:"column:reason" json:"reason"`            // 0序幕礼盒,1免费次数抽中,2邀请次数抽中
	DrawScore   int64  `gorm:"column:draw_score" json:"drawScore"`     // 抽中分数
	Ctime       int64  `gorm:"column:ctime" json:"ctime"`              // 抽奖时间(时间戳秒)
	ScoreBefore int64  `gorm:"column:score_before" json:"scoreBefore"` // 抽中前分数
	Score       int64  `gorm:"column:score" json:"score"`              // 抽中后分数
	ScoreTarget int64  `gorm:"column:score_target" json:"scoreTarget"` // 目标分数
	PrizeType   int32  `gorm:"column:prize_type" json:"prizeType"`     // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	AuditUserid string `gorm:"column:audit_userid" json:"auditUserid"` // 审核人id
	AuditUser   string `gorm:"column:audit_user" json:"auditUser"`     // 审核人
	LuckyDraw   int32  `gorm:"column:lucky_draw" json:"luckyDraw"`     // 1.触发有效邀请直接抽满概率 2.触发有效邀请直接抽满概率且直接抽满
}

func (*ActivityTurnDrawLog) TableName() string {
	return "col_activity_turn_draw_log"
}

func (*ActivityTurnDrawLog) New() CkEntity {
	return new(ActivityTurnDrawLog)
}

func (c *ActivityTurnDrawLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.ActivityTurnDrawLog)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 转盘活动领奖记录
type ActivityTurnPrizeLog struct {
	Ver               int64  `gorm:"column:ver" json:"ver"`                                // 插入时间戳
	Id                string `gorm:"column:id" json:"id"`                                  // 抽奖记录id
	TurnStime         int64  `gorm:"column:turn_stime" json:"turnStime"`                   // 本轮转盘开始时间
	TurnEtime         int64  `gorm:"column:turn_etime" json:"turnEtime"`                   // 本轮转盘结束时间
	Userid            string `gorm:"column:userid" json:"userid"`                          // 用户id
	ChannelId         string `gorm:"column:channel_id" json:"channelId"`                   // 用户渠道
	RegistArea        int8   `gorm:"column:regist_area" json:"registArea"`                 // 用户ab测试 0:A 1:B 2:C
	Rtime             int64  `gorm:"column:rtime" json:"rtime"`                            // 用户注册时间(时间戳秒)
	DrawedTimesFree   int32  `gorm:"column:drawed_times_free" json:"drawedTimesFree" `     // 已使用免费抽奖次数
	DrawedTimesInvite int32  `gorm:"column:drawed_times_invite" json:"drawedTimesInvite" ` // 已使用邀请抽奖次数
	Gives             int64  `gorm:"column:gives" json:"gives"`                            // 赠送序幕金
	Prize             int64  `gorm:"column:prize" json:"prize"`                            // 奖金金额
	PrizeType         int32  `gorm:"column:prize_type" json:"prizeType"`                   // 奖金类型:（1.bonus,2.cash,3.withdrawalble）
	Ctime             int64  `gorm:"column:ctime" json:"ctime"`                            // 申请领奖时间(时间戳秒)
	State             int32  `gorm:"column:state" json:"state"`                            // 审核状态: 0审核中,1通过,2拒绝
	Stime             int64  `gorm:"column:stime" json:"stime"`                            // 审核时间(时间戳秒)
	Stype             int32  `gorm:"column:stype" json:"stype"`                            // 审核类型:1机审,2人工审核
	Reason            string `gorm:"column:reason" json:"reason"`                          // 机审结果原因
	AuditUser         string `gorm:"column:audit_user" json:"auditUser"`                   // 审核人
	Remark            string `gorm:"column:remark" json:"remark"`                          // 后台备注
}

func (*ActivityTurnPrizeLog) TableName() string {
	return "col_activity_turn_prize_log"
}

func (*ActivityTurnPrizeLog) New() CkEntity {
	return new(ActivityTurnPrizeLog)
}

func (c *ActivityTurnPrizeLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.ActivityTurnPrizeLog)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// todo 转盘活动和代理活动的裂变新增人数 裂变归因？
// 代理 打码/人头 奖励记录表 BettingCommission/ReferralBonus
type ShareAgentIncomeRecord struct {
	Ver        int64  `gorm:"column:ver" json:"ver"`                // 插入时间戳
	Id         string `gorm:"column:id" json:"id"`                  // id
	Userid     string `gorm:"column:userid" json:"userid"`          // 用户id
	SuperId    string `gorm:"column:super_id" json:"superId"`       // 上级id
	Lv         int32  `gorm:"column:lv" json:"lv"`                  // 属于几级代理
	TeamLv     int32  `gorm:"column:team_lv" json:"team_lv"`        // 获得奖励的上级团队等级
	Itype      int32  `gorm:"column:itype" json:"itype"`            // 1.打码奖励,2.人数人头奖励,3.累计任务人头奖励
	Amount     int64  `gorm:"column:amount" json:"amount"`          // 返佣奖励金额(毫:1分=10厘=100毫)
	AmountType int32  `gorm:"column:amount_type" json:"amountType"` // 返佣奖励类型: 1:bonus,2:cash,3:withdrawable
	WaterId    string `gorm:"column:water_id" json:"waterId"`       // 打码对局id/人头订单id
	Gtype      int32  `gorm:"column:gtype" json:"gtype"`            // 打码奖励游戏类型
	Bets       int64  `gorm:"column:bets" json:"bets"`              // 打码奖励打码量
	Idate      string `gorm:"column:idate" json:"idate"`            // 日期字符串 2025-01-01
	Ctime      int64  `gorm:"column:ctime" json:"ctime"`            // 创建时间戳毫秒
}

func (*ShareAgentIncomeRecord) TableName() string {
	return "col_activity_share_income_record"
}

func (*ShareAgentIncomeRecord) New() CkEntity {
	return new(ShareAgentIncomeRecord)
}

func (c *ShareAgentIncomeRecord) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.ShareAgentIncomeRecord)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 代理 打码/人头 奖励领取记录表
type ShareAgentIncomeRecordTackLog struct {
	Ver        int64  `gorm:"column:ver" json:"ver"`                // 插入时间戳
	Id         string `gorm:"column:id" json:"id"`                  // id
	SuperId    string `gorm:"column:super_id" json:"superId"`       // 用户id
	Amount     int64  `gorm:"column:amount" json:"amount"`          // 领取奖励金额(分)
	AmountType int32  `gorm:"column:amount_type" json:"amountType"` // 返佣奖励类型: 1:bonus,2:cash,3:withdrawable
	Idate      string `gorm:"column:idate" json:"idate"`            // 日期字符串 2025-01-01
	Itype      int32  `gorm:"column:itype" json:"itype"`            // 1.打码奖励,2人数人头奖励,3.累计任务人头奖励
	Ctime      int64  `gorm:"column:ctime" json:"ctime"`            // 创建时间戳毫秒
}

func (*ShareAgentIncomeRecordTackLog) TableName() string {
	return "col_activity_share_income_record_tack_log"
}

func (*ShareAgentIncomeRecordTackLog) New() CkEntity {
	return new(ShareAgentIncomeRecordTackLog)
}

func (c *ShareAgentIncomeRecordTackLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.ShareAgentIncomeRecordTackLog)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 波动返水记录
type VolatilitySubsidy struct {
	Ver         int64  `gorm:"column:ver" json:"ver"`                  // 插入时间戳
	Id          string `gorm:"column:id" json:"id"`                    // id
	Userid      string `gorm:"column:userid" json:"userid"`            // 用户id
	First       bool   `gorm:"column:first" json:"first"`              // 是否首次领
	FirstTime   int64  `gorm:"column:first_time" json:"firstTime"`     // 首次领取时间戳毫秒
	Sdate       string `gorm:"column:sdate" json:"sdate"`              // 日期字符串 2025-01-01
	Subsidy     int64  `gorm:"column:subsidy" json:"subsidy"`          // 补贴金额
	SubsidyType int32  `gorm:"column:subsidy_type" json:"subsidyType"` // 补贴金额类型: 1:bonus,2:cash,3:withdrawable
	Ctime       int64  `gorm:"column:ctime" json:"ctime"`              // 创建时间戳毫秒
	RegistArea  int32  `gorm:"column:regist_area" json:"registArea"`   // abc类
	Pays        int64  `gorm:"column:pays" json:"pays"`                // 领取时充值
	Bets        int64  `gorm:"column:bets" json:"bets"`                // 领取时打码量
}

func (*VolatilitySubsidy) TableName() string {
	return "col_activity_volatility_subsidy"
}

func (*VolatilitySubsidy) New() CkEntity {
	return new(VolatilitySubsidy)
}

func (c *VolatilitySubsidy) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.VolatilitySubsidy)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 累充转盘解锁记录
type CumRechargeWheelUnlockLog struct {
	Ver    int64  `gorm:"column:ver" json:"ver"`       // 插入时间戳
	Id     string `gorm:"column:id" json:"id"`         // id
	Userid string `gorm:"column:userid" json:"userid"` // 用户id
	Amount int64  `gorm:"column:amount" json:"amount"` // 充值金额
	Times  int32  `gorm:"column:times" json:"times"`   // 解锁次数
	Ctime  int64  `gorm:"column:ctime" json:"ctime"`   // 创建时间
}

func (*CumRechargeWheelUnlockLog) TableName() string {
	return "col_activity_cum_recharge_wheel_unlock_log"
}

func (*CumRechargeWheelUnlockLog) New() CkEntity {
	return new(CumRechargeWheelUnlockLog)
}

func (c *CumRechargeWheelUnlockLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.CumRechargeWheelUnlockLog)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}
	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 累充转盘抽奖记录
type CumRechargeWheelSpinLog struct {
	Ver          int64  `gorm:"column:ver" json:"ver"`                    // 插入时间戳
	Id           string `gorm:"column:id" json:"id"`                      // id
	Userid       string `gorm:"column:userid" json:"userid"`              // 用户id
	Rewardid     int32  `gorm:"column:rewardid" json:"rewardid"`          // 奖项id
	WheelType    int32  `gorm:"column:wheel_type" json:"wheelType"`       // 转盘类型
	RewardType   int32  `gorm:"column:reward_type" json:"rewardType"`     // 奖项类型
	RewardText   string `gorm:"column:reward_text" json:"rewardText"`     // 奖项文本
	RewardAmount int32  `gorm:"column:reward_amount" json:"rewardAmount"` // 奖品数量
	RewardValue  int64  `gorm:"column:reward_value" json:"rewardValue"`   // 奖品价值
	DistType     int32  `gorm:"column:dist_type" json:"distType"`         // 发放类型: 1代表bonus，2代表cash，3代表withdrawable
	Ctime        int64  `gorm:"column:ctime" json:"ctime"`                // 创建时间
}

func (*CumRechargeWheelSpinLog) TableName() string {
	return "col_activity_cum_recharge_wheel_spin_log"
}

func (*CumRechargeWheelSpinLog) New() CkEntity {
	return new(CumRechargeWheelSpinLog)
}

func (c *CumRechargeWheelSpinLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.CumRechargeWheelSpinLog)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}
	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}

// 用户优惠券
type UserCoupon struct {
	Ver         int64  `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id          string `json:"id" gorm:"column:id"`
	UserId      string `json:"user_id" gorm:"column:user_id"`
	CouponId    string `json:"coupon_id" gorm:"column:coupon_id"`       // 优惠券id
	IsAuto      bool   `json:"is_auto" gorm:"column:is_auto"`           // 是否自动
	Amount      int64  `json:"amount" gorm:"column:amount"`             // 优惠金额
	MinRecharge int64  `json:"min_recharge" gorm:"column:min_recharge"` // 最低充值
	Status      bool   `json:"status" gorm:"column:status"`             // 使用状态
	OverTime    int64  `json:"over_time" gorm:"column:over_time"`       // 过期时间
	Ctime       int64  `json:"ctime" gorm:"column:ctime"`               // 创建时间
	Utime       int64  `json:"utime" gorm:"column:utime"`               // 使用时间
}

func (*UserCoupon) TableName() string {
	return "col_user_coupon"
}

func (*UserCoupon) New() CkEntity {
	return new(UserCoupon)
}

func (c *UserCoupon) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.UserCoupon)
	if !ok {
		return nil, fmt.Errorf("parse type %T to %T error", d, c)
	}

	bytes, err := sonic.Marshal(from)
	if err != nil {
		return
	}
	if err = sonic.Unmarshal(bytes, c); err != nil {
		return
	}
	c.Ver = time.Now().UnixMilli()
	rs = append(rs, c)
	return
}
