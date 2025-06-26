package ck

import (
	"fmt"
	"goserver/pkg/data"
	"time"

	"github.com/bytedance/sonic"
)

// 流水日志
type LogWater struct {
	Ver           int64     `gorm:"column:ver" json:"ver"`                         // 插入时间戳
	Id            string    `gorm:"column:id" json:"_id"`                          //id
	Userid        string    `gorm:"column:userid" json:"userid"`                   //账户ID
	Name          string    `gorm:"column:name" json:"name"`                       //名称
	MediaSource   string    `gorm:"column:media_source" json:"media_source"`       //渠道
	WaterDesc     string    `gorm:"column:water_desc" json:"water_desc"`           //流水产生描述
	LType         int32     `gorm:"column:ltype" json:"ltype"`                     //type
	AddDiamond    int64     `gorm:"column:add_diamond" json:"add_diamond"`         //增加彩金
	AddCoin       int64     `gorm:"column:add_coin" json:"add_coin"`               //增加奖励金
	AddOtherAsset int64     `gorm:"column:add_other_asset" json:"add_other_asset"` //增加其他资产
	ChangeAsset   int64     `gorm:"column:change_asset" json:"change_asset"`       //变化总资产
	OldDiamond    int64     `gorm:"column:old_diamond" json:"old_diamond"`         //变化前彩金
	NowDiamond    int64     `gorm:"column:now_diamond" json:"now_diamond"`         //当前彩金
	OldCoin       int64     `gorm:"column:old_coin" json:"old_coin"`               //变化前奖励金
	NowCoin       int64     `gorm:"column:now_coin" json:"now_coin"`               //当前奖励金
	OldOtherAsset int64     `gorm:"column:old_other_asset" json:"old_other_asset"` //变化前其他资产
	NowOtherAsset int64     `gorm:"column:now_other_asset" json:"now_other_asset"` //当前其他资产
	WaterId       string    `gorm:"column:water_id" json:"water_id"`               //流水号
	Ctime         time.Time `gorm:"column:ctime" json:"ctime"`                     //流水产生时间
	Control       bool      `gorm:"column:control" json:"control"`                 //点控局
}

func (*LogWater) TableName() string {
	return "col_log_water"
}

func (*LogWater) New() CkEntity {
	return new(LogWater)
}

func (c *LogWater) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.LogWater)
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

// 登录日志
type LogLogin struct {
	//Id         string `gorm:"column:_id"`
	Ver        int64     `gorm:"column:ver" json:"ver"`                 // 插入时间戳
	Userid     string    `gorm:"column:userid" json:"userid"`           //账户ID
	Event      int       `gorm:"column:event" json:"event"`             //事件：0=登录,1=正常退出,2＝系统关闭时被迫退出,3＝被动退出,4＝其它情况导致的退出
	Ip         string    `gorm:"column:ip" json:"ip"`                   //登录IP
	DayStamp   time.Time `gorm:"column:day_stamp" json:"day_stamp"`     //login Time Today
	LoginTime  time.Time `gorm:"column:login_time" json:"login_time"`   //login Time
	LogoutTime time.Time `gorm:"column:logout_time" json:"logout_time"` //logout Time
	Atype      uint32    `gorm:"column:atype" json:"atype"`             //login type
}

func (*LogLogin) TableName() string {
	return "col_log_login"
}

func (*LogLogin) New() CkEntity {
	return new(LogLogin)
}

func (c *LogLogin) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.LogLogin)
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

type NsqLogExternalBet struct {
	Ver             int64  `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id              string `json:"_id" gorm:"column:id"`
	UserId          string `json:"user_id" gorm:"column:user_id"`
	GameId          int32  `json:"game_id" gorm:"column:game_id"`
	RoundId         string `json:"round_id" gorm:"column:round_id"`
	MerchantOrderNo string `json:"merchant_order_no" gorm:"column:merchant_order_no"`
	OrderNo         string `json:"order_no" gorm:"column:order_no"`
	Amount          int64  `json:"amount" gorm:"column:amount"`
	Ctime           int64  `json:"ctime" gorm:"column:ctime"`
	Cancel          bool   `json:"cancel" gorm:"column:cancel"`
}

func (*NsqLogExternalBet) TableName() string {
	return "col_nsq_log_external_bet"
}

func (*NsqLogExternalBet) New() CkEntity {
	return new(NsqLogExternalBet)
}

func (c *NsqLogExternalBet) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.NsqLogExternalBet)
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

type NsqLogExternalReward struct {
	Ver             int64  `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id              string `json:"_id" gorm:"column:id"`
	UserId          string `json:"user_id" gorm:"column:user_id"`
	GameId          int32  `json:"game_id" gorm:"column:game_id"`
	RoundId         string `json:"round_id" gorm:"column:round_id"`
	MerchantOrderNo string `json:"merchant_order_no" gorm:"column:merchant_order_no"`
	OrderNo         string `json:"order_no" gorm:"column:order_no"`
	RewardAmount    int64  `json:"amount" gorm:"column:amount"`
	Ctime           int64  `json:"ctime" gorm:"column:ctime"`
	Cancel          bool   `json:"cancel" gorm:"column:cancel"`
}

func (*NsqLogExternalReward) TableName() string {
	return "col_nsq_log_external_reward"
}

func (*NsqLogExternalReward) New() CkEntity {
	return new(NsqLogExternalReward)
}

func (c *NsqLogExternalReward) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.NsqLogExternalReward)
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

type NsqLogExternalCancel struct {
	Ver               int64  `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id                string `json:"_id" gorm:"column:id"`
	Type              int32  `json:"type" gorm:"column:type"`
	UserId            string `json:"user_id" gorm:"column:user_id"`
	GameId            int32  `json:"game_id" gorm:"column:game_id"`
	RoundId           string `json:"round_id" gorm:"column:round_id"`
	MerchantOrderNo   string `json:"merchant_order_no" gorm:"column:merchant_order_no"`
	OrderNo           string `json:"order_no" gorm:"column:order_no"`
	CancelOrderNo     string `json:"cancel_order_no" gorm:"column:cancel_order_no"`
	CancelPlatOrderNo string `json:"cancel_plat_order_no" gorm:"column:cancel_plat_order_no"`
	OrderAmount       int64  `json:"order_amount" gorm:"column:order_amount"`
	OrderDesc         string `json:"order_desc" gorm:"column:order_desc"`
	Ctime             int64  `json:"ctime" gorm:"column:ctime"`
}

func (*NsqLogExternalCancel) TableName() string {
	return "col_nsq_log_external_cancel"
}

func (*NsqLogExternalCancel) New() CkEntity {
	return new(NsqLogExternalCancel)
}

func (c *NsqLogExternalCancel) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.NsqLogExternalCancel)
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

// LaunchInviteLog 发起邀请日志(点邀请按钮)
type LaunchInviteLog struct {
	Ver        int64  `gorm:"column:ver" json:"ver"` // 插入时间戳
	Id         string `gorm:"column:id" json:"id"`
	Userid     string `gorm:"column:userid" json:"userid"`
	Ltype      int8   `gorm:"column:ltype" json:"ltype"` // 1主界面分享,2转盘分享,3代理活动界面分享
	ChannelId  string `gorm:"column:channel_id" json:"channelId"`
	RegistArea int8   `gorm:"column:regist_area" json:"registArea"`
	Rtime      int64  `gorm:"column:rtime" json:"rtime"` // 用户注册时间(毫秒)
	Ctime      int64  `gorm:"column:ctime" json:"ctime"` // 时间戳(毫秒)
}

func (*LaunchInviteLog) TableName() string {
	return "col_launch_invite_log"
}

func (*LaunchInviteLog) New() CkEntity {
	return new(LaunchInviteLog)
}

func (c *LaunchInviteLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.LaunchInviteLog)
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

// vbbank变化日志
type LogVbDiamond struct {
	Ver           int64  `gorm:"column:ver" json:"ver"`                       // 插入时间戳
	Id            string `json:"id" gorm:"column:id"`                         //id
	Userid        string `json:"userid" gorm:"column:userid"`                 //账户ID
	Name          string `json:"name" gorm:"column:name"`                     //名称
	MediaSource   string `json:"mediaSource" gorm:"column:media_source"`      //渠道
	GenPosition   string `json:"genPosition" gorm:"column:gen_position"`      //产生位置
	GameId        string `json:"gameId" gorm:"column:game_id"`                //局号/订单号
	BeforeOutCash int64  `json:"beforeOutCash" gorm:"column:before_out_cash"` //账变前可提现金额
	AfterOutCash  int64  `json:"afterOutCash" gorm:"column:after_out_cash"`   //账变后可提现金额
	Vb            int64  `json:"vb" gorm:"column:vb"`                         //账变值
	Ctime         int64  `json:"ctime" gorm:"column:ctime"`                   //时间(秒)
	Reason        int32  `json:"reason" gorm:"column:reason"`                 //原因
}

func (*LogVbDiamond) TableName() string {
	return "col_log_vbdiamond"
}

func (*LogVbDiamond) New() CkEntity {
	return new(LogVbDiamond)
}

func (c *LogVbDiamond) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.LogVbDiamond)
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
	c.Vb = c.AfterOutCash - c.BeforeOutCash
	rs = append(rs, c)
	return
}

// 可提现金变化日志
type LogOutDiamond struct {
	Ver           int64  `gorm:"column:ver" json:"ver"`                       // 插入时间戳
	Id            string `json:"id" gorm:"column:id"`                         //id
	Userid        string `json:"userid" gorm:"column:userid"`                 //账户ID
	Name          string `json:"name" gorm:"column:name"`                     //名称
	MediaSource   string `json:"mediaSource" gorm:"column:media_source"`      //渠道
	GenPosition   string `json:"genPosition" gorm:"column:gen_position"`      //产生位置
	GameId        string `json:"gameId" gorm:"column:game_id"`                //局号/订单号
	BeforeOutCash int64  `json:"beforeOutCash" gorm:"column:before_out_cash"` //账变前可提现金额
	AfterOutCash  int64  `json:"afterOutCash" gorm:"column:after_out_cash"`   //账变后可提现金额
	OutCash       int64  `json:"outCash" gorm:"column:out_cash"`              //账变值
	Ctime         int64  `json:"ctime" gorm:"column:ctime"`                   //时间(秒)
	Reason        int32  `json:"reason" gorm:"column:reason"`                 //原因
}

func (*LogOutDiamond) TableName() string {
	return "col_log_outdiamond"
}

func (*LogOutDiamond) New() CkEntity {
	return new(LogOutDiamond)
}

func (c *LogOutDiamond) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.LogOutDiamond)
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
	c.OutCash = c.AfterOutCash - c.BeforeOutCash
	rs = append(rs, c)
	return
}

// 每分钟在线人数记录
type OnlineUsersLog struct {
	Ver             int64  `gorm:"column:ver" json:"ver"`                           // 插入时间戳
	Id              int64  `gorm:"column:id" json:"id"`                             // id
	Ctime           int64  `gorm:"column:ctime" json:"ctime"`                       // 记录时间
	DayTime         uint32 `gorm:"column:day_time" json:"dayTime"`                  // 年月日 20250422
	MinuteTime      uint32 `gorm:"column:minute_time" json:"minuteTime"`            // 时分 1611
	Users           uint32 `gorm:"column:users" json:"users"`                       // 玩家人数
	PayUsers        uint32 `gorm:"column:pay_users" json:"payUsers"`                // 付费玩家人数
	OnlineUsers     uint32 `gorm:"column:online_users" json:"onlineUsers"`          // 在线玩家人数
	OnlinePayUsers  uint32 `gorm:"column:online_pay_users" json:"onlinePayUsers"`   // 在线付费玩家人数
	OfflineUsers    uint32 `gorm:"column:offline_users" json:"offlineUsers"`        // 离线玩家人数(可能在外接)
	OfflinePayUsers uint32 `gorm:"column:offline_pay_users" json:"offlinePayUsers"` // 离线付费玩家人数(可能在外接)
}

func (*OnlineUsersLog) TableName() string {
	return "col_log_online_users"
}

func (*OnlineUsersLog) New() CkEntity {
	return new(OnlineUsersLog)
}

func (c *OnlineUsersLog) ParseData(d interface{}) (rs []interface{}, err error) {
	// check type
	from, ok := d.(*data.OnlineUsersLog)
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
