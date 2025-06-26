package data

import (
	"encoding/json"
	"fmt"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

//TODO 数据统计 玩家7日，30日，总赢亏

// 1注册赠送,2开房消耗,3房间解散返还,
// 4充值购买,5下注,7上庄，8下庄
// 8下庄, 9后台操作,11破产补助
// 18商城购买,19绑定赠送,20首充赠送
// 23进入房间消耗
// 24通比牛牛,25看牌抢庄,26牛牛抢庄
// 27代理发放,28vip赠送
// 38退款,39本金返还,40输赢,41坐庄输赢
// 42异常退款,43反佣
// 44机器人破产补助
// 45庄家抽佣
const (
	LogType1  int32 = 1
	LogType2  int32 = 2
	LogType3  int32 = 3
	LogType4  int32 = 4
	LogType5  int32 = 5
	LogType7  int32 = 7
	LogType8  int32 = 8
	LogType9  int32 = 9
	LogType11 int32 = 11
	LogType18 int32 = 18
	LogType19 int32 = 19
	LogType20 int32 = 20
	LogType23 int32 = 23
	LogType24 int32 = 24
	LogType25 int32 = 25
	LogType26 int32 = 26
	LogType27 int32 = 27
	LogType28 int32 = 28
	LogType38 int32 = 38
	LogType39 int32 = 39
	LogType40 int32 = 40
	LogType41 int32 = 41
	LogType42 int32 = 42
	LogType43 int32 = 43
	LogType44 int32 = 44
	LogType45 int32 = 45
)

// 注册日志
type LogRegist struct {
	//Id       string    `bson:"_id"`
	Userid   string    `bson:"userid"`    //账户ID
	Nickname string    `bson:"nickname"`  //账户名称
	Ip       string    `bson:"ip"`        //注册IP
	DayStamp time.Time `bson:"day_stamp"` //regist Time Today
	DayDate  int       `bson:"day_date"`  //regist day date
	Ctime    time.Time `bson:"ctime"`     //create Time
	Atype    uint32    `bson:"atype"`     //regist type
}

func (this *LogRegist) Save() bool {
	//this.Id = bson.NewObjectId().String()
	// this.DayStamp = utils.TimestampTodayTime()
	this.DayDate = utils.DayDate()
	this.Ctime = bson.Now()
	return Insert(LogRegists, this)
}

// 注册记录
func RegistRecord(userid, nickname, ip string, atype uint32, loc *time.Location) {
	record := &LogRegist{
		Userid:   userid,
		Nickname: nickname,
		Ip:       ip,
		Atype:    atype,
		DayStamp: utils.TimestampTodayTime(loc),
	}
	record.Save()
}

// 登录日志
type LogLogin struct {
	//Id         string `bson:"_id"`
	Userid     string    `bson:"userid" json:"userid"`           //账户ID
	Event      int       `bson:"event" json:"event"`             //事件：0=登录,1=正常退出,2＝系统关闭时被迫退出,3＝被动退出,4＝其它情况导致的退出
	Ip         string    `bson:"ip" json:"ip"`                   //登录IP
	DayStamp   time.Time `bson:"day_stamp" json:"day_stamp"`     //login Time Today
	LoginTime  time.Time `bson:"login_time" json:"login_time"`   //login Time
	LogoutTime time.Time `bson:"logout_time" json:"logout_time"` //logout Time
	Atype      uint32    `bson:"atype" json:"atype"`             //login type
}

func (this *LogLogin) Save() bool {
	//this.Id = bson.NewObjectId().String()
	// this.DayStamp = utils.TimestampTodayTime()
	this.LoginTime = bson.Now()
	return Insert(LogLogins, this)
}

func (this *LogLogin) Update(event int) bool {
	this.LogoutTime = bson.Now()
	return Update(LogLogins, bson.M{"userid": this.Userid, "event": 0},
		bson.M{"$set": bson.M{"event": event, "logout_time": this.LogoutTime}})
}

func (this *LogLogin) FindUpdateBefore() {
	GetByQ(LogLogins, bson.M{"userid": this.Userid, "event": 0}, this)
}

// 登录记录
func LoginRecord(userid, ip string, atype uint32, loc *time.Location) {
	record := &LogLogin{
		Userid:   userid,
		Event:    0,
		Ip:       ip,
		Atype:    atype,
		DayStamp: utils.TimestampTodayTime(loc),
	}
	record.Save()

	DataProducers.PublishDatas(mq.TopicSyncLogLogin, []any{record})
}

// 登录记录
func LogoutRecord(userid string, event int) {
	record := &LogLogin{
		Userid: userid,
	}
	record.FindUpdateBefore()
	record.Update(event)

	record.Event = event
	DataProducers.PublishDatas(mq.TopicSyncLogLogin, []any{record})
}

// 流水日志
type LogWater struct {
	Id            string    `bson:"_id" json:"_id"`                         //id
	Userid        string    `bson:"userid" json:"userid"`                   //账户ID
	Name          string    `bson:"name" json:"name"`                       //名称
	MediaSource   string    `bson:"media_source" json:"media_source"`       //渠道
	WaterDesc     string    `bson:"water_desc" json:"water_desc"`           //流水产生描述
	LType         int32     `bson:"ltype" json:"ltype"`                     //type
	AddDiamond    int64     `bson:"add_diamond" json:"add_diamond"`         //增加彩金
	AddCoin       int64     `bson:"add_coin" json:"add_coin"`               //增加奖励金
	AddOtherAsset int64     `bson:"add_other_asset" json:"add_other_asset"` //增加其他资产
	ChangeAsset   int64     `bson:"change_asset" json:"change_asset"`       //变化总资产
	OldDiamond    int64     `bson:"old_diamond" json:"old_diamond"`         //变化前彩金
	NowDiamond    int64     `bson:"now_diamond" json:"now_diamond"`         //当前彩金
	OldCoin       int64     `bson:"old_coin" json:"old_coin"`               //变化前奖励金
	NowCoin       int64     `bson:"now_coin" json:"now_coin"`               //当前奖励金
	OldOtherAsset int64     `bson:"old_other_asset" json:"old_other_asset"` //变化前其他资产
	NowOtherAsset int64     `bson:"now_other_asset" json:"now_other_asset"` //当前其他资产
	WaterId       string    `bson:"water_id" json:"water_id"`               //流水号
	Ctime         time.Time `bson:"ctime" json:"ctime"`                     //流水产生时间
	Control       bool      `bson:"control" json:"control"`                 //点控局
}

// 可提现彩金变化日志
type LogOutDiamond struct {
	Id            string `json:"id" bson:"_id"`                        //id
	Userid        string `json:"userid" bson:"userid"`                 //账户ID
	Name          string `json:"name" bson:"name"`                     //名称
	MediaSource   string `json:"mediaSource" bson:"media_source"`      //渠道
	GenPosition   string `json:"genPosition" bson:"gen_position"`      //产生位置
	GameId        string `json:"gameId" bson:"game_id"`                //局号/订单号
	BeforeOutCash int64  `json:"beforeOutCash" bson:"before_out_cash"` //账变前可提现金额
	AfterOutCash  int64  `json:"afterOutCash" bson:"after_out_cash"`   //账变后可提现金额
	BeforeCash    int64  `json:"beforeCash" bson:"before_cash"`        //账变前彩金
	AfterCash     int64  `json:"afterCash" bson:"after_cash"`          //账变后彩金
	Ctime         int64  `json:"ctime" bson:"ctime"`                   //时间
	Reason        int32  `json:"reason" bson:"reason"`                 //原因
}

// vbbank变化日志
type LogVbDiamond struct {
	Id            string `json:"id" bson:"_id"`                        //id
	Userid        string `json:"userid" bson:"userid"`                 //账户ID
	Name          string `json:"name" bson:"name"`                     //名称
	MediaSource   string `json:"mediaSource" bson:"media_source"`      //渠道
	GenPosition   string `json:"genPosition" bson:"gen_position"`      //产生位置
	GameId        string `json:"gameId" bson:"game_id"`                //局号/订单号
	BeforeOutCash int64  `json:"beforeOutCash" bson:"before_out_cash"` //账变前可提现金额
	AfterOutCash  int64  `json:"afterOutCash" bson:"after_out_cash"`   //账变后可提现金额
	Ctime         int64  `json:"ctime" bson:"ctime"`                   //时间
	Reason        int32  `json:"reason" bson:"reason"`                 //原因
}

// 钻石日志
type LogDiamond struct {
	//Id     string `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Type   int32     `bson:"type"`   //类型
	Num    int64     `bson:"num"`    //数量
	Rest   int64     `bson:"rest"`   //剩余数量
	Ctime  time.Time `bson:"ctime"`  //create Time
}

// 分享提取流水
type LogShareWithDraw struct {
	Id       string `json:"_id" bson:"_id"`
	Userid   string `json:"userid" bson:"userid"`
	BundleId string `json:"bundleId" bson:"bundle_id"`
	Gtype    int32  `json:"gtype" bson:"gtype"`
	Cash     int64  `json:"cash" bson:"cash"`
	Coin     int64  `json:"coin" bson:"coin"`
	Ctime    int64  `json:"ctime" bson:"ctime"`
}

type LogEventTrack struct {
	Id     string `json:"_id" bson:"_id"`
	Typ    int32  `json:"typ" bson:"typ"`
	Userid string `json:"userid" bson:"userid"`
	Gtype  int32  `json:"gtype" bson:"gtype"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

type LogGameTime struct {
	Id     string `json:"_id" bson:"_id"`
	Userid string `json:"userid" bson:"userid"`
	Gtype  int32  `json:"gtype" bson:"gtype"`
	Time   int64  `json:"time" bson:"time"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

type LogButtonClick struct {
	Id     string `json:"_id" bson:"_id"`
	Userid string `json:"userid" bson:"userid"`
	Type   string `json:"type" bson:"type"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

type LogButton2Click struct {
	Id     string `json:"_id" bson:"_id"`
	Userid string `json:"userid" bson:"userid"`
	Type   string `json:"type" bson:"type"`
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

type LogADReport struct {
	Id         string `json:"_id" bson:"_id"`
	ADID       string `json:"adid" bson:"adid"`
	BundleID   string `json:"bundle_id" bson:"bundle_id"`
	Userid     string `json:"userid" bson:"userid"`
	Type       string `json:"type" bson:"type"`
	Code       int    `json:"code" bson:"code"`
	Etime      int64  `json:"etime" bson:"etime"` //事件时间
	Ctime      int64  `json:"ctime" bson:"ctime"`
	StatusCode int    `json:"status_code" bson:"status_code"`
	Status     string `json:"status" bson:"status"`
}

type NsqLogExternalBet struct {
	Id              string `json:"_id" bson:"_id"`
	UserId          string `json:"user_id" bson:"user_id"`
	GameId          int32  `json:"game_id" bson:"game_id"`
	RoundId         string `json:"round_id" bson:"round_id"`
	MerchantOrderNo string `json:"merchant_order_no" bson:"merchant_order_no"`
	OrderNo         string `json:"order_no" bson:"order_no"`
	Amount          int64  `json:"amount" bson:"amount"`
	Ctime           int64  `json:"ctime" bson:"ctime"`
	Cancel          bool   `json:"cancel" bson:"cancel"`
}

type NsqLogExternalReward struct {
	Id              string `json:"_id" bson:"_id"`
	UserId          string `json:"user_id" bson:"user_id"`
	GameId          int32  `json:"game_id" bson:"game_id"`
	RoundId         string `json:"round_id" bson:"round_id"`
	MerchantOrderNo string `json:"merchant_order_no" bson:"merchant_order_no"`
	OrderNo         string `json:"order_no" bson:"order_no"`
	RewardAmount    int64  `json:"amount" bson:"amount"`
	Ctime           int64  `json:"ctime" bson:"ctime"`
	Cancel          bool   `json:"cancel" bson:"cancel"`
}

type NsqLogExternalCancel struct {
	Id                string `json:"_id" bson:"_id"`
	Type              int32  `json:"type" bson:"type"`
	UserId            string `json:"user_id" bson:"user_id"`
	GameId            int32  `json:"game_id" bson:"game_id"`
	RoundId           string `json:"round_id" bson:"round_id"`
	MerchantOrderNo   string `json:"merchant_order_no" bson:"merchant_order_no"`
	OrderNo           string `json:"order_no" bson:"order_no"`
	CancelOrderNo     string `json:"cancel_order_no" bson:"cancel_order_no"`
	CancelPlatOrderNo string `json:"cancel_plat_order_no" bson:"cancel_plat_order_no"`
	OrderAmount       int64  `json:"order_amount" bson:"order_amount"`
	OrderDesc         string `json:"order_desc" bson:"order_desc"`
	Ctime             int64  `json:"ctime" bson:"ctime"`
}

type LogRechargeReport struct {
	Id      string    `json:"_id" bson:"_id"`
	Uid     string    `json:"uid" bson:"uid"`
	Code    int32     `json:"code" bson:"code"`
	Content string    `json:"content" bson:"content"`
	Ctime   time.Time `json:"ctime" bson:"ctime"`
}

type LogFBReport struct {
	Id         string `json:"_id" bson:"_id"`
	AppId      string `json:"appId" bson:"app_id"`
	Userid     string `json:"userid" bson:"userid"`
	Type       string `json:"type" bson:"type"`
	Code       int    `json:"code" bson:"code"`
	Etime      int64  `json:"etime" bson:"etime"` //事件时间
	Ctime      int64  `json:"ctime" bson:"ctime"`
	StatusCode int    `json:"status_code" bson:"status_code"`
	Status     string `json:"status" bson:"status"`
	ReqParam   string `json:"req_param" bson:"req_param"`
}

type NsqLogEfiTransaction struct {
	MemberName   string `json:"MemberName" bson:"member_name"`     // 运营商中玩家的唯一标识符
	OperatorCode string `json:"OperatorCode" bson:"operator_code"` // Seamless中运营商的唯一标识符,BO登录用户名
	MessageID    string `json:"MessageID" bson:"message_id"`       // 当前API请求的唯一标识符
	RequestTime  string `json:"RequestTime" bson:"request_time"`   // 请求日期时间 日期时间格式为yyyyMMddHHmmss
	TransType    int32  `json:"TransType" bson:"trans_type"`       // 交易类型

	TransactionID     string    `json:"TransactionID" bson:"_id"`                    // 当前交易的唯一标识符,请返回错误代码1003以表示检测到重复交易。
	MemberID          int64     `json:"MemberID" bson:"member_id"`                   // Seamless中玩家的唯一标识符
	OperatorID        int64     `json:"OperatorID" bson:"operator_id"`               // Seamless中运营商的唯一标识符,有时它被称为AgentID
	ProductID         int64     `json:"ProductID" bson:"product_id"`                 // Seamless中产品的唯一标识符
	ProviderID        int32     `json:"ProviderID" bson:"provider_id"`               // Seamless 中供应商的唯一标识符。
	ProviderLineID    int32     `json:"ProviderLineID" bson:"provider_line_id"`      // Seamless 中对产品线配置的唯一标识符。
	WagerID           int64     `json:"WagerID" bson:"wager_id"`                     // Seamless 投注记录的唯一标识符
	CurrencyID        int32     `json:"CurrencyID" bson:"currency_id"`               // Seamless中货币的唯一标识符 INR=16
	GameType          int32     `json:"GameType" bson:"game_type"`                   // 交易的游戏类型
	GameID            string    `json:"GameID" bson:"game_id"`                       // 供应商游戏代码
	GameRoundID       string    `json:"GameRoundID" bson:"game_round_id"`            // 供应商游戏轮次ID
	ValidBetAmount    float64   `json:"ValidBetAmount" bson:"valid_bet_amount"`      // 在扣除营业额赢利后的投注金额
	BetAmount         float64   `json:"BetAmount" bson:"bet_amount"`                 // 整个下注金额，未扣除营业额中奖金额
	TransactionAmount float64   `json:"TransactionAmount" bson:"transaction_amount"` // 需要对玩家钱包进行更改的金额, 正值表示增加玩家钱包金额，负值表示减少玩家钱包金额
	PayoutAmount      float64   `json:"PayoutAmount" bson:"payout_amount"`           // 玩家的赢取金额，如果玩家输了，这个值可以为0。
	PayoutDetail      string    `json:"PayoutDetail" bson:"payout_detail"`           // 由供应商发送的有关玩家赢取的详细信息。
	CommissionAmount  float64   `json:"CommissionAmount" bson:"commission_amount"`   // 佣金金额
	JackpotAmount     float64   `json:"JackpotAmount" bson:"jackpot_amount"`         // 奖池金额
	SettlementDate    string    `json:"SettlementDate" bson:"settlement_date"`       // 最终确认投注记录的日期,当投注结束时，玩家要么输了要么赢了。
	JPBet             float64   `json:"JPBet" bson:"jp_bet"`                         // 供应商每次下注的奖池贡献金额,另一个术语是渐进式累计奖池
	Status            int32     `json:"Status" bson:"status"`                        // 当前交易的状态 100 Pending, 101 Settle, 102 Void
	CreatedOn         string    `json:"CreatedOn" bson:"created_on"`                 // 交易日期
	ModifiedOn        string    `json:"ModifiedOn" bson:"modified_on"`               // 修改交易的日期
	Ctime             time.Time `json:"ctime" bson:"ctime"`
}

func (this *LogDiamond) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogDiamonds, this)
}

func (this *LogWater) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogWaters, this)
}

func (l *LogOutDiamond) SaveOutCash() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncLogOutDiamond, []any{l})

	//this.Id = bson.NewObjectId().String()
	l.Ctime = bson.Now().Unix()
	return Insert(LogOutDiamonds, l)
}

func (l *LogOutDiamond) SaveGiveCash() bool {
	//this.Id = bson.NewObjectId().String()
	l.Ctime = bson.Now().Unix()
	return Insert(LogGiveDiamonds, l)
}

func (l *LogOutDiamond) SaveVbLog() bool {
	l.Ctime = bson.Now().Unix()

	defer DataProducers.PublishDatas(mq.TopicSyncLogVBDiamond, []any{&LogVbDiamond{
		Id:            l.Id,
		Userid:        l.Userid,
		Name:          l.Name,
		MediaSource:   l.MediaSource,
		GenPosition:   l.GenPosition,
		GameId:        l.GameId,
		BeforeOutCash: l.BeforeOutCash,
		AfterOutCash:  l.AfterOutCash,
		Ctime:         l.Ctime,
		Reason:        l.Reason,
	}})
	return Insert(LogVBDiamonds, l)
}

func (l *LogShareWithDraw) Save() bool {
	//this.Id = bson.NewObjectId().String()
	l.Ctime = bson.Now().Unix()
	return Insert(LogShareWaters, l)
}

func (l *LogEventTrack) Save() bool {
	return Insert(LogEventTracks, l)
}

func (l *LogGameTime) Save() bool {
	return Insert(LogGameTimes, l)
}

func (l *LogButtonClick) Save() bool {
	return Insert(LogButtonClicks, l)
}

func (l *LogButton2Click) Save() bool {
	return Insert(LogButton2Clicks, l)
}

func (l *LogADReport) Save() bool {
	return Insert(LogADReports, l)
}

func (l *NsqLogExternalBet) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncExternalBet, []any{l})

	return Insert(NsqLogExternalBets, l)
}

func (l *LogFBReport) Save() bool {
	return Insert(LogFBReports, l)
}

func (l *NsqLogExternalBet) Get(merchant_order_no string, order_no string) {
	GetByQ(NsqLogExternalBets, bson.M{
		"user_id":           l.UserId,
		"game_id":           l.GameId,
		"round_id":          l.RoundId,
		"merchant_order_no": merchant_order_no,
		"order_no":          order_no,
		"cancel":            false,
	}, l)
}

func (l *NsqLogExternalBet) CancelBet() {
	defer func() {
		l.Cancel = false
		DataProducers.PublishDatas(mq.TopicSyncExternalBet, []any{l})
	}()

	Update(NsqLogExternalBets, bson.M{
		"user_id":           l.UserId,
		"game_id":           l.GameId,
		"round_id":          l.RoundId,
		"merchant_order_no": l.MerchantOrderNo,
		"order_no":          l.OrderNo,
		"cancel":            false,
	}, bson.M{"$set": bson.M{"cancel": true}})
}

func (l *NsqLogExternalReward) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncExternalReward, []any{l})

	return Insert(NsqLogExternalRewards, l)
}

func (l *NsqLogExternalReward) Get(merchant_order_no string, order_no string) {
	GetByQ(NsqLogExternalRewards, bson.M{
		"user_id":           l.UserId,
		"game_id":           l.GameId,
		"round_id":          l.RoundId,
		"merchant_order_no": merchant_order_no,
		"order_no":          order_no,
		"cancel":            false,
	}, l)
}

func (l *NsqLogExternalReward) CancelReward() {
	defer DataProducers.PublishDatas(mq.TopicSyncExternalReward, []any{l})

	Update(NsqLogExternalRewards, bson.M{
		"user_id":           l.UserId,
		"game_id":           l.GameId,
		"round_id":          l.RoundId,
		"merchant_order_no": l.MerchantOrderNo,
		"order_no":          l.OrderNo,
		"cancel":            false,
	}, bson.M{"$set": bson.M{"cancel": true}})
}

func (l *NsqLogExternalCancel) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncExternalCancel, []any{l})

	return Insert(NsqLogExternalCancels, l)
}

func (l *NsqLogEfiTransaction) Save() bool {
	return Insert(NsqLogEfiTransactions, l)
}

func (l *LogRechargeReport) Save() bool {
	return Insert(LogRechargeReports, l)
}

func EventTrack(arg *pb.LogEventTrack) {
	record := &LogEventTrack{
		Id:     bson.NewObjectId().String(),
		Typ:    int32(arg.Typ),
		Userid: arg.Userid,
		Gtype:  arg.Gtype,
		Ctime:  bson.Now().Unix(),
	}
	record.Save()
}

func GameTime(arg *pb.LogGameTime) {
	record := &LogGameTime{
		Id:     bson.NewObjectId().String(),
		Userid: arg.Userid,
		Gtype:  arg.Gtype,
		Time:   arg.Time,
		Ctime:  bson.Now().Unix(),
	}
	record.Save()
}

func ButtonClick(arg *pb.LogButtonClick) {
	record := &LogButtonClick{
		Id:     bson.NewObjectId().String(),
		Userid: arg.Userid,
		Type:   arg.Typ,
		Ctime:  bson.Now().Unix(),
	}
	record.Save()

	ButtonClick2(&pb.LogButtonClick2{
		Userid: arg.Userid,
		Typ:    arg.Typ,
	})
}

// 全部按钮点击记录
func ButtonClick2(arg *pb.LogButtonClick2) {
	record := &LogButton2Click{
		Id:     bson.NewObjectId().String(),
		Userid: arg.Userid,
		Type:   arg.Typ,
		Ctime:  bson.Now().Unix(),
	}
	record.Save()
}

func RechargeReport(arg *pb.LogRechargeReport) {
	record := &LogRechargeReport{
		Id:      bson.NewObjectId().String(),
		Uid:     arg.Uid,
		Code:    arg.Code,
		Content: arg.Content,
		Ctime:   utils.LocalTime(),
	}
	record.Save()
}

// func ADReport(adid, bundleid, userid, typ string, code int) {
// 	record := &LogADReport{
// 		Id:       bson.NewObjectId().String(),
// 		ADID:     adid,
// 		BundleID: bundleid,
// 		Userid:   userid,
// 		Type:     typ,
// 		Code:     code,
// 		Ctime:    bson.Now().Unix(),
// 	}
// 	record.Save()
// }

func ADReport2(userid, bundleid, adid, typ string, etime int64, status_code int, status string) {
	var code int
	if status_code == 200 {
		code = 0
	} else {
		code = -1
	}

	record := &LogADReport{
		Id:         bson.NewObjectId().String(),
		ADID:       adid,
		BundleID:   bundleid,
		Userid:     userid,
		Type:       typ,
		Code:       code,
		Etime:      etime,
		Ctime:      bson.Now().Unix(),
		StatusCode: status_code,
		Status:     status,
	}
	record.Save()
}

func LogExternalBet(body *pb.NsqLogExternalBet) {
	log := &NsqLogExternalBet{
		Id:              bson.NewObjectId().String(),
		UserId:          body.Uid,
		GameId:          body.GameId,
		RoundId:         body.RoundId,
		MerchantOrderNo: body.MerchantOrderNo,
		OrderNo:         body.OrderNo,
		Amount:          body.Amount,
		Ctime:           bson.Now().Unix(),
	}
	log.Save()
}

func LogExternalReward(body *pb.NsqLogExternalReward) {
	log := &NsqLogExternalReward{
		Id:              bson.NewObjectId().String(),
		UserId:          body.Uid,
		GameId:          body.GameId,
		RoundId:         body.RoundId,
		MerchantOrderNo: body.MerchantOrderNo,
		OrderNo:         body.OrderNo,
		RewardAmount:    body.RewardAmount,
		Ctime:           bson.Now().Unix(),
	}
	log.Save()
}

func LogExternalCancel(body *pb.NsqLogExternalCancel) {
	log := &NsqLogExternalCancel{
		Id:                bson.NewObjectId().String(),
		Type:              body.Type,
		UserId:            body.Uid,
		GameId:            body.GameId,
		RoundId:           body.RoundId,
		MerchantOrderNo:   body.MerchantOrderNo,
		OrderNo:           body.OrderNo,
		CancelOrderNo:     body.CancelOrderNo,
		CancelPlatOrderNo: body.CancelPlatOrderNo,
		OrderAmount:       body.OrderAmount,
		OrderDesc:         body.OrderDesc,
	}
	log.Save()
}

// 视讯交易
func LogEfiTransaction(body *pb.EfiTransactionReq) {
	for _, trans := range body.Transactions {
		log := &NsqLogEfiTransaction{
			MemberName:        body.MemberName,
			OperatorCode:      body.OperatorCode,
			MessageID:         body.MessageID,
			RequestTime:       body.RequestTime,
			TransType:         int32(body.TransType),
			TransactionID:     trans.TransactionID,
			MemberID:          trans.MemberID,
			OperatorID:        trans.OperatorID,
			ProductID:         trans.ProductID,
			ProviderID:        trans.ProviderID,
			ProviderLineID:    trans.ProviderLineID,
			WagerID:           trans.WagerID,
			CurrencyID:        trans.CurrencyID,
			GameType:          trans.GameType,
			GameID:            trans.GameID,
			GameRoundID:       trans.GameRoundID,
			ValidBetAmount:    trans.ValidBetAmount,
			BetAmount:         trans.BetAmount,
			TransactionAmount: trans.TransactionAmount,
			PayoutAmount:      trans.PayoutAmount,
			PayoutDetail:      trans.PayoutDetail,
			CommissionAmount:  trans.CommissionAmount,
			JackpotAmount:     trans.JackpotAmount,
			SettlementDate:    trans.SettlementDate,
			JPBet:             trans.JPBet,
			Status:            trans.Status,
			CreatedOn:         trans.CreatedOn,
			ModifiedOn:        trans.ModifiedOn,
			Ctime:             time.Now(),
		}
		if log.TransactionID == "" {
			log.TransactionID = bson.NewObjectId().String()
		}
		if !log.Save() {
			glog.Errorf("save LogEfiTransaction: %v\n", log)
		}
	}
}

// 分享日志
func LogShare(arg *pb.LogShareWithDraw) {
	record := &LogShareWithDraw{
		Id:       bson.NewObjectId().String(),
		Userid:   arg.Userid,
		BundleId: arg.Adid,
		Gtype:    arg.Gtype,
		Cash:     arg.Cash,
		Coin:     arg.Coin,
	}
	record.Save()
}

// 流水日志
func CurrencyRecord(arg *pb.LogCurrencyWater) {
	record := &LogWater{
		Id:            bson.NewObjectId().String(),
		Userid:        arg.Userid,
		Name:          arg.Name,
		LType:         arg.Ltype,
		MediaSource:   arg.MediaSource,
		WaterDesc:     arg.WaterDesc,
		AddDiamond:    arg.AddDiamond,
		AddCoin:       arg.AddCoin,
		AddOtherAsset: arg.AddOtherAsset,
		ChangeAsset:   arg.AddDiamond + arg.AddCoin + arg.AddOtherAsset,
		OldDiamond:    arg.Diamond - arg.AddDiamond,
		NowDiamond:    arg.Diamond,
		OldCoin:       arg.Coin - arg.AddCoin,
		NowCoin:       arg.Coin,
		OldOtherAsset: arg.OtherAsset - arg.AddOtherAsset,
		NowOtherAsset: arg.OtherAsset,
		WaterId:       arg.WaterId,
		Control:       arg.Control,
	}
	record.Save()

	DataProducers.PublishDatas(mq.TopicSyncLogWater, []any{record})
}

// 可提现彩金日志
func OutDiamondLog(arg *pb.LogWithdrawCash) {
	record := &LogOutDiamond{
		Id:            bson.NewObjectId().String(),
		Userid:        arg.Userid,
		Name:          arg.Name,
		MediaSource:   arg.MediaSource,
		GenPosition:   arg.GenPosition,
		GameId:        arg.GameId,
		BeforeOutCash: arg.BeforeOutCash,
		AfterOutCash:  arg.AfterOutCash,
		BeforeCash:    arg.BeforeCash,
		AfterCash:     arg.AfterCash,
		Reason:        arg.Reason,
	}
	switch arg.Dtype {
	case 1:
		record.SaveOutCash()
	case 2:
		record.SaveGiveCash()
	case 3:
		record.SaveVbLog()
	}
}

// 钻石记录
func DiamondRecord(userid string, rtype int32, rest, num int64) {
	record := &LogDiamond{
		Userid: userid,
		Type:   rtype,
		Num:    num,
		Rest:   rest,
	}
	record.Save()
}

// TODO 添加索引优化查询
// err := collection.EnsureIndex(index)
// 金币日志
type LogCoin struct {
	//Id     string `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Type   int32     `bson:"type"`   //类型
	Num    int64     `bson:"num"`    //数量
	Rest   int64     `bson:"rest"`   //剩余数量
	Ctime  time.Time `bson:"ctime"`  //create Time
}

func (this *LogCoin) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogCoins, this)
}

// 金币记录
func CoinRecord(userid string, rtype int32, rest, num int64) {
	record := &LogCoin{
		Userid: userid,
		Type:   rtype,
		Num:    num,
		Rest:   rest,
	}
	record.Save()
}

// 房卡日志
type LogCard struct {
	//Id     string `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Type   int32     `bson:"type"`   //类型
	Num    int64     `bson:"num"`    //数量
	Rest   int64     `bson:"rest"`   //剩余数量
	Ctime  time.Time `bson:"ctime"`  //create Time
}

func (this *LogCard) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogCards, this)
}

// 房卡记录
func CardRecord(userid string, rtype int32, rest, num int64) {
	record := &LogCard{
		Userid: userid,
		Type:   rtype,
		Num:    num,
		Rest:   rest,
	}
	record.Save()
}

// 筹码日志
type LogChip struct {
	//Id     string `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Type   int32     `bson:"type"`   //类型
	Num    int64     `bson:"num"`    //数量
	Rest   int64     `bson:"rest"`   //剩余数量
	Ctime  time.Time `bson:"ctime"`  //create Time
}

func (this *LogChip) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogChips, this)
}

// 筹码记录
func ChipRecord(userid string, rtype int32, rest, num int64) {
	record := &LogChip{
		Userid: userid,
		Type:   rtype,
		Num:    num,
		Rest:   rest,
	}
	record.Save()
}

// 绑定日志
type LogBuildAgency struct {
	//Id       string `bson:"_id"`
	Userid    string    `bson:"userid"`     //账户ID
	Agent     string    `bson:"agent"`      //绑定ID
	DayStamp  time.Time `bson:"day_stamp"`  //regist Time Today
	Day       int       `bson:"day"`        //regist day
	Month     int       `bson:"month"`      //regist month
	Ctime     time.Time `bson:"ctime"`      //create Time
	DayDate   int       `bson:"day_date"`   //regist day date
	MonthDate int       `bson:"month_date"` //regist month date
}

func (this *LogBuildAgency) Save() bool {
	//this.Id = bson.NewObjectId().Hex()
	// this.DayStamp = utils.TimestampTodayTime()
	this.DayDate = utils.DayDate()
	this.MonthDate = utils.MonthDate()
	this.Day = utils.Day()
	this.Month = int(utils.Month())
	this.Ctime = bson.Now()
	return Insert(LogBuildAgencys, this)
}

// 绑定记录
func BuildRecord(userid, agent string, loc *time.Location) {
	record := &LogBuildAgency{
		Userid:   userid,
		Agent:    agent,
		DayStamp: utils.TimestampTodayTime(loc),
	}
	record.Save()
}

// 在线日志
type LogOnline struct {
	//Id       string `bson:"_id"`
	Num      int       `bson:"num"`       //online count
	DayStamp time.Time `bson:"day_stamp"` //Time Today
	Ctime    time.Time `bson:"ctime"`     //create Time
}

func (this *LogOnline) Save() bool {
	//this.Id = bson.NewObjectId().Hex()
	// this.DayStamp = utils.TimestampTodayTime()
	this.Ctime = bson.Now()
	return Insert(LogOnlines, this)
}

// 在线记录
func OnlineRecord(num int, loc *time.Location) {
	record := &LogOnline{
		Num:      num,
		DayStamp: utils.TimestampTodayTime(loc),
	}
	record.Save()
}

// 期号日志
type LogExpect struct {
	//Id     string `bson:"_id"`
	Expect    string    `bson:"expect"`     //期号ID
	Codes     string    `bson:"codes"`      //开奖号码
	OpenTimer int64     `bson:"open_timer"` //开奖时间
	Ctime     time.Time `bson:"ctime"`      //create Time
}

func (this *LogExpect) Save() bool {
	//this.Id = bson.NewObjectId().String()
	this.Ctime = bson.Now()
	return Insert(LogExpects, this)
}

// 期号记录
func ExpectRecord(expect, codes string, openTimer int64) {
	record := &LogExpect{
		Expect:    expect,
		Codes:     codes,
		OpenTimer: openTimer,
	}
	record.Save()
}

// LogTask 任务日志
type LogTask struct {
	//Id       string    `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Taskid int32     `bson:"taskid"` //taskid
	Type   int32     `bson:"type"`   //task type
	Ctime  time.Time `bson:"ctime"`  //create Time
}

// Save 保存消息记录
func (t *LogTask) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Ctime = bson.Now()
	return Insert(LogTasks, t)
}

// TaskRecord 任务记录
func TaskRecord(userid string, taskid, ttype int32) {
	record := &LogTask{
		Userid: userid,
		Taskid: taskid,
		Type:   ttype,
	}
	record.Save()
}

// LogProfit 代理收益日志
type LogProfit struct {
	//Id       string    `bson:"_id"`
	Agentid string    `bson:"agentid"` //代理ID,to
	Userid  string    `bson:"userid"`  //玩家ID,from
	Gtype   int32     `bson:"gtype"`   //game type
	Level   uint32    `bson:"level"`   //level type, 表示相对agentid的等级
	Rate    uint32    `bson:"rate"`    //rate
	Profit  int64     `bson:"profit"`  //Profit
	Type    int32     `bson:"type"`    //类型
	Ctime   time.Time `bson:"ctime"`   //create Time
}

// Save 保存消息记录
func (t *LogProfit) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Ctime = bson.Now()
	return Insert(LogProfits, t)
}

// ProfitRecord 代理收益记录
func ProfitRecord(arg *pb.LogProfit) {
	record := &LogProfit{
		Userid:  arg.GetUserid(),
		Agentid: arg.GetAgentid(),
		Gtype:   arg.GetGtype(),
		Level:   arg.GetLevel(),
		Rate:    arg.GetRate(),
		Profit:  arg.GetProfit(),
		Type:    arg.GetType(),
	}
	record.Save()
	//添加天统计
	DayProfitRecord(arg)
}

// LogDayProfit 代理收益统计日志
type LogDayProfit struct {
	//Id       string    `bson:"_id"`
	Agentid      string    `bson:"agentid"`       //代理ID,to
	Userid       string    `bson:"userid"`        //玩家ID,from
	Nickname     string    `bson:"nickname"`      //昵称,from
	AgentNote    string    `bson:"agent_note"`    // 代理备注,from
	Day          int       `bson:"day"`           //day
	Profit       int64     `bson:"profit"`        //Profit
	ProfitFirst  int64     `bson:"profit_first"`  //ProfitFirst
	ProfitSecond int64     `bson:"profit_second"` //ProfitSecond
	ProfitMonth  int64     `bson:"profit_month"`  //ProfitMonth
	Utime        time.Time `bson:"utime"`         //update Time
	Ctime        time.Time `bson:"ctime"`         //create Time
}

// Save 保存消息记录
func (t *LogDayProfit) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Ctime = bson.Now()
	return Insert(LogDayProfits, t)
}

// Has 记录是否存在
func (t *LogDayProfit) Has() bool {
	return Has(LogDayProfits, bson.M{"userid": t.Userid, "agentid": t.Agentid, "day": t.Day})
}

// Update 更新记录
func (t *LogDayProfit) Update(field string, profit int64) bool {
	if field == "" {
		return false
	}
	return Upsert(LogDayProfits, bson.M{"userid": t.Userid, "agentid": t.Agentid, "day": t.Day},
		bson.M{"$set": bson.M{"utime": t.Utime, "agent_note": t.AgentNote, "nickname": t.Nickname, "ctime": t.Utime},
			"$inc": bson.M{field: profit}})
}

// DayProfitRecord 代理收益统计记录
func DayProfitRecord(arg *pb.LogProfit) {
	record := &LogDayProfit{
		Userid:  arg.GetUserid(),
		Agentid: arg.GetAgentid(),
		//Profit: arg.GetProfit(),
		Nickname:  arg.GetNickname(),
		AgentNote: arg.GetAgentnote(),
	}
	record.Utime = bson.Now()
	record.Day = utils.Time2DayDate(record.Utime)
	var field string
	switch arg.GetType() {
	case int32(pb.LOG_TYPE54): //区域奖励发放不记录
	case int32(pb.LOG_TYPE53):
		field = "profit_month"
		record.ProfitMonth = arg.GetProfit()
	case int32(pb.LOG_TYPE52):
		switch arg.GetLevel() {
		case 1:
			field = "profit"
			record.Profit = arg.GetProfit()
		case 2:
			field = "profit_first"
			record.ProfitFirst = arg.GetProfit()
		case 3:
			field = "profit_second"
			record.ProfitSecond = arg.GetProfit()
		}
	}
	record.Update(field, arg.GetProfit())
	//if record.Has() {
	//	record.Update(field)
	//} else {
	//	record.Save()
	//}
}

// LogSysProfit 系统收益日志
type LogSysProfit struct {
	//Id       string    `bson:"_id"`
	Agentid string    `bson:"agentid"` //代理ID
	Userid  string    `bson:"userid"`  //玩家ID
	Gtype   int32     `bson:"gtype"`   //game type
	Level   uint32    `bson:"level"`   //level type, 表示userid等级
	Rate    uint32    `bson:"rate"`    //rate
	Profit  int64     `bson:"profit"`  //Profit
	Rest    int64     `bson:"rest"`    //Rest
	Ctime   time.Time `bson:"ctime"`   //create Time
}

// Save 保存消息记录
func (t *LogSysProfit) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Ctime = bson.Now()
	return Insert(LogSysProfits, t)
}

// SysProfitRecord 系统收益记录
func SysProfitRecord(agentid, userid string, gtype int32, level, rate uint32, profit, rest int64) {
	record := &LogSysProfit{
		Userid:  userid,
		Agentid: agentid,
		Gtype:   gtype,
		Level:   level,
		Rate:    rate,
		Profit:  profit,
		Rest:    rest,
	}
	record.Save()
}

// LogBank 银行日志
type LogBank struct {
	//Id     string `bson:"_id"`
	Userid string    `bson:"userid"` //账户ID
	Type   int32     `bson:"type"`   //类型
	Num    int64     `bson:"num"`    //数量
	Rest   int64     `bson:"rest"`   //剩余数量
	From   string    `bson:"from"`   //赠送者
	Ctime  time.Time `bson:"ctime"`  //create Time
}

func (t *LogBank) Save() bool {
	//t.Id = bson.NewObjectId().String()
	t.Ctime = bson.Now()
	return Insert(LogBanks, t)
}

// 银行记录
func BankRecord(userid, from string, rtype int32, rest, num int64) {
	record := &LogBank{
		Userid: userid,
		Type:   rtype,
		Num:    num,
		Rest:   rest,
		From:   from,
	}
	record.Save()
}

// GetBankLogs 获取银行操作记录
// func GetBankLogs(arg *pb.BankLogReq) ([]LogBank, error) {
// 	if arg.Page == 0 {
// 		arg.Page = 1
// 	}
// 	pageSize := 20 //取前20条
// 	skipNum, sortFieldR := parsePageAndSort(int(arg.Page), pageSize, "ctime", false)
// 	var list []LogBank
// 	q := bson.M{"userid": arg.Userid}
// 	err := LogBanks.
// 		Find(q).
// 		Sort(sortFieldR).
// 		Skip(skipNum).
// 		Limit(pageSize).
// 		All(&list)
// 	if err != nil {
// 	}
// 	return list, err
// }

// bug日志
type BugFeedbackLog struct {
	Id     string  `json:"id" bson:"_id"`         // id
	UserId string  `json:"userId" bson:"user_id"` // 用户id
	BugId  []int32 `json:"bugId" bson:"bug_id"`   // bugid
	Ctime  int64   `json:"ctime" bson:"ctime"`    // 提交时间
}

func (b *BugFeedbackLog) Save() {
	b.Id = bson.NewObjectId().String()
	b.Ctime = utils.LocalTime().UnixMilli()
	Insert(LogBugFeedbacks, b)
}

func BugLog(arg *pb.LogBugFeedback) {
	log := &BugFeedbackLog{
		UserId: arg.Userid,
		BugId:  arg.Bugid,
	}
	log.Save()
}

func FeedBackReq(arg *pb.FeedBackReq) {
	if arg.ReqId != 0 {
		content := ""
		switch arg.ReqId {
		case 1:
			content = "Please share us the amount you paid successfully."
		case 2:
			content = "Please share us the exact time you paid successfully."
		case 3:
			// content = "Please share us the UTR code."
			content = "Please upload payment screenshot with UTR code in your game deposit order record."
		case 4:
			content = "Please wait, we are checking with bank.The amount will be added into your game account automatically once bank gets the payment."
		}

		if content != "" {
			log := ChatLog{Sender: "-1", Content: content, Ctime: utils.BsonNow(), Name: "system"}
			if Has(ChatLogs, bson.M{"_id": arg.UserId}) {
				Update(ChatLogs, bson.M{"_id": arg.UserId}, bson.M{"$push": bson.M{"log": log}, "$set": bson.M{"utime": utils.BsonNow()}})
			} else {
				logs := make([]ChatLog, 0)
				logs = append(logs, log)
				msg := FeedBackLog{
					Userid:      arg.UserId,
					Logs:        logs,
					ReplyStatus: 3,
					Utime:       utils.BsonNow(),
				}
				Insert(ChatLogs, msg)
			}
		}
	}

	log := ChatLog{Sender: arg.UserId, Content: arg.Content, Ctime: utils.BsonNow(), Name: arg.Name}

	if Has(ChatLogs, bson.M{"_id": arg.UserId}) {
		Update(ChatLogs, bson.M{"_id": arg.UserId}, bson.M{"$push": bson.M{"log": log}, "$set": bson.M{"utime": utils.BsonNow(), "reply_status": 1}})
		return
	}

	logs := make([]ChatLog, 0)
	logs = append(logs, log)
	msg := FeedBackLog{
		Userid:      arg.UserId,
		Logs:        logs,
		ReplyStatus: 1,
		Utime:       utils.BsonNow(),
	}
	Insert(ChatLogs, msg)
	// log.Get()
	// log.Logs = append(log.Logs, *chat)
	// log.Utime = utils.BsonNow()

	// log.Save()
}

func LogSmsRecordReq(arg *pb.LogSmsRecord) {
	log := new(LogSmsRecord)
	err := json.Unmarshal(arg.Body, &log)
	if err != nil {
		return
	}
	if log.Id == "" {
		return
	}

	log.Ctime = utils.BsonNow().Unix()
	log.Save()
}

func LogUserSmsRecord(arg *pb.LogUserSmsRecord) {
	log := &LogUserSms{
		Id:    bson.NewObjectId().String(),
		Phone: fmt.Sprint("0091", arg.Phone),
		Code:  arg.Code,
		Ctime: utils.BsonNow().Unix(),
	}
	log.Save()
}

type LogScratchTicket struct {
	Id        string `json:"id" bson:"_id"` // id
	Userid    string `json:"userid" bson:"userid"`
	Win       bool   `json:"win" bson:"win"`              //是否中奖
	Jackpot1  int64  `json:"jackpot1" bson:"jackpot1"`    //一等奖
	Jackpot2  int64  `json:"jackpot2" bson:"jackpot2"`    //二等奖
	Jackpot3  int64  `json:"jackpot3" bson:"jackpot3"`    //三等奖
	Jackpot4  int64  `json:"jackpot4" bson:"jackpot4"`    //四等奖
	Jackpot5  int64  `json:"jackpot5" bson:"jackpot5"`    //五等奖
	CostBonus int64  `json:"costBonus" bson:"cost_bonus"` //消耗bonus
	Ctime     int64  `json:"ctime" bson:"ctime"`
}

func LogScratchTicketData(arg *pb.ScratchTicketLog) {
	log := &LogScratchTicket{
		Id:        bson.NewObjectId().String(),
		Userid:    arg.Userid,
		CostBonus: arg.Cost,
		Win:       len(arg.Level) > 0,
		Ctime:     utils.BsonNow().Unix(),
	}

	for i, v := range arg.Level {
		switch v {
		case 1:
			log.Jackpot1 = arg.Bonus[i]
		case 2:
			log.Jackpot2 = arg.Bonus[i]
		case 3:
			log.Jackpot3 = arg.Bonus[i]
		case 4:
			log.Jackpot4 = arg.Bonus[i]
		case 5:
			log.Jackpot5 = arg.Bonus[i]
		}
	}
	log.Save()
}

func (log *LogScratchTicket) Save() {
	Insert(LogScratchTickets, log)
}

type LogActivityMonitor struct {
	Id     string `json:"id" bson:"_id"` // id
	Userid string `json:"userid" bson:"userid"`
	Otype  int32  `json:"otype" bson:"otype"`
	FuncId int32  `json:"funcId" bson:"func_id"` // 功能id
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

func ActivityMonitorLog(arg *pb.ActivityMonitorReq) {
	log := &LogActivityMonitor{
		Id:     bson.NewObjectId().String(),
		Userid: arg.Userid,
		Otype:  arg.Otype,
		FuncId: arg.Funcid,
		Ctime:  utils.BsonNow().Unix(),
	}
	log.Save()
}

func (log *LogActivityMonitor) Save() {
	Insert(LogActivityMonitors, log)
}

type LogPlayShareDraw struct {
	Id      string `json:"id" bson:"_id"` // id
	Userid  string `json:"userid" bson:"userid"`
	Dtype   int32  `json:"dtype" bson:"dtype"` // 1玩游戏 2分享 3领取奖励
	Jackpot int64  `json:"jackpot" bson:"jackpot"`
	Ctime   int64  `json:"ctime" bson:"ctime"`
}

func PlayShareDrawLog(arg *pb.PlayShareDrawLog) {
	log := &LogPlayShareDraw{
		Id:      bson.NewObjectId().String(),
		Userid:  arg.Userid,
		Dtype:   arg.Dtype,
		Jackpot: arg.Reward,
		Ctime:   utils.BsonNow().Unix(),
	}
	log.Save()
}

func (log *LogPlayShareDraw) Save() {
	Insert(LogPlayShareDraws, log)
}

type LogBonus struct {
	Id     string `json:"id" bson:"_id"` // id
	Userid string `json:"userid" bson:"userid"`
	Bonus  int64  `json:"bonus" bson:"bonus"` // bonus变化
	Ctime  int64  `json:"ctime" bson:"ctime"`
}

func BonusLog(arg *pb.BonusLog) {
	log := &LogBonus{
		Id:     bson.NewObjectId().String(),
		Userid: arg.Userid,
		Bonus:  arg.Bonus,
		Ctime:  utils.BsonNow().Unix(),
	}
	log.Save()
}

func (log *LogBonus) Save() {
	Insert(LogBonuss, log)
}

type LogShareRegist struct {
	Userid   string `json:"userid" bson:"_id"`
	Superior string `json:"superior" bson:"superior"`
	Ctime    int64  `json:"ctime" bson:"ctime"`
}

func ShareRegistLog(userid, superior string) {
	log := &LogShareRegist{
		Userid:   userid,
		Superior: superior,
		Ctime:    utils.BsonNow().Unix(),
	}
	log.Save()
}

func (log *LogShareRegist) Save() {
	Insert(LogShareRegists, log)
}

type LogGameFlowWater struct {
	Userid string                        `json:"userid" bson:"_id"`
	Detail map[int][]*FlowWaterDetailLog `json:"detail" bson:"detail"`
	Ctime  int64                         `json:"ctime" bson:"ctime"`
}

type FlowWaterDetailLog struct {
	Value   int64 `json:"value" bson:"value"`
	Gtype   int32 `json:"gtype" bson:"gtype"`
	Outside int32 `json:"outside" bson:"outside"` // 1：slots 2：live
}

func GameFlowWaterLog(arg *pb.GameFlowWaterLog) {
	log := &LogGameFlowWater{
		Userid: arg.Userid,
		Ctime:  utils.BsonNow().Unix(),
	}
	detail := &FlowWaterDetailLog{
		Value:   arg.Val,
		Gtype:   arg.Gtype,
		Outside: arg.Outside,
	}

	log.Update(detail, int(arg.ChargeTimes))
}

func (log *LogGameFlowWater) Update(detail *FlowWaterDetailLog, times int) {
	d := make(map[int][]*FlowWaterDetailLog)
	if Has(LogGameFlowWaters, bson.M{"_id": log.Userid}) {
		tempLog := new(LogGameFlowWater)
		GetByQWithFields(LogGameFlowWaters, bson.M{"_id": log.Userid}, []string{"detail"}, &tempLog)
		if tempLog != nil {
			d = tempLog.Detail
		}
		if v, ok := d[times]; ok {
			add := true
			for _, l := range v {
				if l.Gtype == detail.Gtype && l.Outside == detail.Outside {
					add = false
					l.Value += detail.Value
					break
				}
			}
			if add {
				d[times] = append(v, detail)
			}
		} else {
			d[times] = []*FlowWaterDetailLog{detail}
		}
		// 直接更新
		Update(LogGameFlowWaters, bson.M{"_id": log.Userid}, bson.M{"$set": bson.M{"detail": d}})
		return
	}
	// 新增
	d[times] = []*FlowWaterDetailLog{detail}
	log.Detail = d
	Insert(LogGameFlowWaters, log)
}

// 用户自定义头像上传记录
type UploadUserHeadLog struct {
	Id      string `bson:"_id" json:"id"`           // id
	UserId  string `json:"userId" bson:"user_id"`   // 用户ID
	Url     string `json:"url" bson:"url"`          // 头像地址
	IsAudit int64  `json:"isAudit" bson:"is_audit"` // 审核状态 0:待审核 1：通过；2：拒绝
	Ctime   int64  `json:"ctime" bson:"ctime"`      // 创建时间
}

func UploadUserHeadRecord(arg *pb.UploadPhotoLog) {
	log := &UploadUserHeadLog{
		Id:     bson.NewObjectId().String(),
		UserId: arg.Userid,
		Url:    arg.Url,
		Ctime:  utils.BsonNow().UnixMilli(),
	}
	log.Save()
}

func (log *UploadUserHeadLog) Save() {
	Insert(UploadUserHeads, log)
}

func GetPhotoLog(id string) UploadUserHeadLog {
	log := UploadUserHeadLog{}
	Get(UploadUserHeads, id, &log)
	return log
}

type JHVFRecord struct {
	Id         string    `bson:"_id" json:"id"`                  // id ante + cards
	UserId     string    `json:"userId" bson:"user_id"`          // 用户ID
	Ante       int32     `json:"ante" bson:"ante"`               // 房间底分
	Cards      []uint32  `json:"cards" bson:"cards"`             // 牌面
	CardsId    int32     `json:"cards_id" bson:"cards_id"`       // 牌面id
	BetAmount  int64     `json:"bet_amount" bson:"bet_amount"`   // 下注额
	CardsTimes int32     `json:"cards_times" bson:"cards_times"` // 拿到该牌面次数
	Utime      time.Time `json:"utime" bson:"utime"`             // 更新时间
}

func (vf *JHVFRecord) FindById() {
	Get(JHVFRecords, vf.Id, vf)
}

func (vf *JHVFRecord) Save() {
	Upsert(JHVFRecords, bson.M{"_id": vf.Id}, vf)
}

// tp vf 记录
func LogJHVFRecordSync(arg *pb.JHVFRecordSync) {
	// 牌面+底注s
	algo.Sort(arg.Cards)

	var cardsId int32 = int32(arg.Cards[0])<<16 | int32(arg.Cards[1])<<8 | int32(arg.Cards[2])
	var vfid int64 = int64(arg.Ante)<<24 | int64(cardsId)

	vf := &JHVFRecord{
		Id: fmt.Sprintf("%s-%d", arg.Userid, vfid),
	}
	vf.FindById()
	vf.UserId = arg.Userid
	vf.Cards = arg.Cards
	vf.CardsId = cardsId
	vf.Ante = arg.Ante
	vf.BetAmount += int64(arg.BetAmount)
	vf.CardsTimes++
	vf.Utime = time.Now()
	vf.Save()
}

// LaunchInviteLog 发起邀请日志(点邀请按钮)
type LaunchInviteLog struct {
	Id         string `bson:"_id" json:"id"`
	Userid     string `bson:"userid" json:"userid"`
	Ltype      int8   `bson:"ltype" json:"ltype"` // 1主界面分享,2转盘分享
	ChannelId  string `bson:"channel_id" json:"channelId"`
	RegistArea int8   `bson:"regist_area" json:"registArea"`
	Rtime      int64  `bson:"rtime" json:"rtime"` // 用户注册时间毫秒
	Ctime      int64  `bson:"ctime" json:"ctime"` // 时间戳毫秒
}

func (l *LaunchInviteLog) Save() {
	DataProducers.PublishDatas(mq.TopicSyncLaunchInviteLog, []any{l})
}

func LogLaunchInvite(arg *pb.LaunchInviteLog) {
	log := &LaunchInviteLog{
		Id:         bson.NewObjectId().Hex(),
		Userid:     arg.Userid,
		Ltype:      int8(arg.Ltype),
		ChannelId:  arg.ChannelId,
		RegistArea: int8(arg.RegistArea),
		Rtime:      arg.Rtime,
		Ctime:      arg.Ctime,
	}
	log.Save()
}

// 每分钟在线人数记录
type OnlineUsersLog struct {
	Id              int64  `bson:"_id" json:"id"`                            // id
	Ctime           int64  `json:"ctime" bson:"ctime"`                       // 记录时间
	DayTime         uint32 `json:"dayTime" bson:"day_time"`                  // 年月日 20250422
	MinuteTime      uint32 `json:"minuteTime" bson:"minute_time"`            // 时分 1611
	Users           uint32 `json:"users" bson:"users"`                       // 玩家人数
	PayUsers        uint32 `json:"payUsers" bson:"pay_users"`                // 付费玩家人数
	OnlineUsers     uint32 `json:"onlineUsers" bson:"online_users"`          // 在线玩家人数
	OnlinePayUsers  uint32 `json:"onlinePayUsers" bson:"online_pay_users"`   // 在线付费玩家人数
	OfflineUsers    uint32 `json:"offlineUsers" bson:"offline_users"`        // 离线玩家人数(可能在外接)
	OfflinePayUsers uint32 `json:"offlinePayUsers" bson:"offline_pay_users"` // 离线付费玩家人数(可能在外接)
}

// Save 写入数据库
func (t *OnlineUsersLog) Save() bool {
	defer DataProducers.PublishDatas(mq.TopicSyncOnlineUsersLog, []any{t})
	return Upsert(OnlineUsersLogs, bson.M{"_id": t.Id}, t)
}

// Save 写入数据库
func LogOnlineUsersLog(log *pb.OnlineUsersLog) {
	record := &OnlineUsersLog{
		Id:              log.Id,
		Ctime:           log.Ctime,
		DayTime:         log.DayTime,
		MinuteTime:      log.MinuteTime,
		Users:           log.Users,
		PayUsers:        log.PayUsers,
		OnlineUsers:     log.OnlineUsers,
		OnlinePayUsers:  log.OnlinePayUsers,
		OfflineUsers:    log.OfflineUsers,
		OfflinePayUsers: log.OfflinePayUsers,
	}
	record.Save()
}
