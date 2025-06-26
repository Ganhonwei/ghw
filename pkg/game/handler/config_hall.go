package handler

import (
	"bytes"
	"errors"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/game/config"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/dhushon/decxls"
)

// 分享配置
type Share struct {
	Id            int32  `xls:"分享活动ID"`
	Recharge      int32  `xls:"领奖充值要求"`
	FirstRecharge int32  `xls:"下线首单充值奖励"`
	MaxPeople     int32  `xls:"下线人数上限"`
	Url           string `xls:"分享地址"`
	PCUrl         string `xls:"PC分享地址"`
	FirstReward   int32  `xls:"第1层奖励比例（万分比）"`
	SecondReward  int32  `xls:"第2层奖励比例（万分比）"`
}

type ShareAddr struct {
	Id   string `xls:"渠道包名"`
	Link string `xls:"分享链接"`
}

// 新手配置
type Beginner struct {
	Id               int32   `xls:"新手配置id"`
	RegistMode1      int32s  `xls:"模式1（A，B）"`
	RegistMode2      int32s  `xls:"模式2（A，B）"`
	NoviceGame       int32ss `xls:"AB类型引导进入房间（gtype，gameid）"`
	Recharge         int32   `xls:"正常玩家充值金额条件"`
	PlayerGames      int32   `xls:"正常玩家局数要求"`
	OnlineTime       int32   `xls:"正常玩家在线时长要求（分钟）"`
	OutCashInterval  int32s  `xls:"新手可提彩金区间"`
	Winning          int32s  `xls:"对应胜率(万分比）"`
	OutCashLimited   int32   `xls:"新手可提现彩金上限"`
	FrothMaxRecharge int32   `xls:"泡沫检测充值上限（分）"`
	FrothGiftRate    int     `xls:"泡沫赠送率（万分比）"`
	FrothBetInterval ints    `xls:"泡沫下注额度范围"`
	FrothWonRate     ints    `xls:"泡沫下注额度对应胜率"`
	FactorSeed       float64 `xls:"玩家系数种子"`
	NewbiewToCivil   int32s  `xls:"新手转平民（游戏时长min，游戏局数）"`
}

// 房间人数配置
type RoomPeople struct {
	ID              int32  `xls:"id"`
	Gtype           int32  `xls:"游戏ID"`
	RoomId          string `xls:"房间ID"`
	PeopleInterval1 int32s `xls:"闲时上下限"`
	PeopleInterval2 int32s `xls:"正常上下限"`
	PeopleInterval3 int32s `xls:"忙时上下限"`
}

// 周卡
type WeekCard struct {
	ID     int32  `xls:"周卡ID"`
	Price  int32  `xls:"价格（分）"`
	Reward int32  `xls:"奖励"`
	Give   int32s `xls:"每日赠送彩金"`
}

type Emoji struct {
	Id              int32 `xls:"ID"`
	RobotTypeWeight ints  `xls:"人机类型权重"`
	RobotType       ints  `xls:"人机类型"`
}

// 签到
type DailySignIn struct {
	ID       int32 `xls:"第n天"`
	Reward   int32 `xls:"奖励"`
	Recharge int32 `xls:"价格"`
}

// 在线奖励
type OnlineReward struct {
	ID         int32   `xls:"ID"`
	OnlineTime int64   `xls:"在线时长"`
	Reward     int32ss `xls:"奖励"`
}

// 限时礼包
type LimitedGift struct {
	ID         int32 `xls:"ID"`
	Switch     int   `xls:"开关"`
	Price      int32 `xls:"礼包价格"`
	Reward     int64 `xls:"实得彩金"`
	DuringTime int64 `xls:"倒数时间（秒）"`
	Recharge   int32 `xls:"充值金额要求"`
	GameRound  int32 `xls:"游戏局数要求"`
}

// 人机表情配置
type EmojiConfig struct {
	Id                 int  `xls:"人机类型"`
	DeskSendProb       int  `xls:"上桌发送概率"`
	DeskSendWeight     ints `xls:"上桌表情权重"`
	DeskSendDelay      ints `xls:"上桌触发延迟"`
	GameStartProb      int  `xls:"游戏开始发送概率"`
	GameStartWeight    ints `xls:"游戏开始表情权重"`
	GameStartDelay     ints `xls:"游戏开始触发延迟"`
	WaitLongProb       int  `xls:"等待太久发送概率"`
	WaitLongWeight     ints `xls:"等待太久表情权重"`
	WaitLongDelay      ints `xls:"等待太久触发延迟"`
	ReplyProb          int  `xls:"回复表情"`
	ReplyWeight        ints `xls:"回复表情权重"`
	ReplyDelay         ints `xls:"回复触发延迟"`
	OtherSeeProb       int  `xls:"其他玩家看牌后"`
	OtherSeeWeight     ints `xls:"其他玩家看牌后表情权重"`
	OtherSeeDelay      ints `xls:"其他玩家看牌后触发延迟"`
	OtherRaiseProb     int  `xls:"其他玩家加注"`
	OtherRaiseWeight   ints `xls:"其他玩家加注表情权重"`
	OtherRaiseDelay    ints `xls:"其他玩家加注触发延迟"`
	OtherBiProb        ints `xls:"其他玩家比牌"`
	OtherBiWeight      ints `xls:"其他玩家比牌表情权重"`
	OtherBiDelay       ints `xls:"其他玩家比牌触发延迟"`
	SelfBiWinProb      int  `xls:"自己比牌赢"`
	SelfBiWinWeight    ints `xls:"自己比牌赢表情权重"`
	SelfBiWinDelay     ints `xls:"自己比牌赢触发延迟"`
	MoPreCardProb      int  `xls:"摸上家牌"`
	MoPreCardWeight    ints `xls:"摸上家牌表情权重"`
	MoPreCardDelay     ints `xls:"摸上家牌触发延迟"`
	OtherChuCardProb   int  `xls:"其他玩家出牌"`
	OtherChuCardWeight ints `xls:"其他玩家出牌表情权重"`
	OtherChuCardDelay  ints `xls:"其他玩家出牌触发延迟"`
}

// 对战房基本配置表
type PvpRoom struct {
	Id             int   `xls:"ID"`
	PvpChargeLimit int   `xls:"开启功能充值要求"`
	GameStartWait  int   `xls:"房间等待时间秒"`
	TpFunInitScore int   `xls:"TP娱乐模式初始分"`
	RmFunInitScore int   `xls:"RM娱乐模式初始分"`
	AbFunInitScore int   `xls:"AB娱乐模式初始分"`
	TpGoldRound    ints  `xls:"TP真金模式局数"`
	RmGoldRound    ints  `xls:"RM真金模式局数"`
	AbGoldRound    ints  `xls:"AB真金模式局数"`
	TpFunRoundCost intss `xls:"TP娱乐模式局数和费用"`
	RmFunRoundCost intss `xls:"RM娱乐模式局数和费用"`
	AbFunRoundCost intss `xls:"AB娱乐模式局数和费用"`
}

// vip
type VIP struct {
	Level         int   `xls:"VIP等级"`
	NextLevel     int   `xls:"下一等级"`
	DailySign     int64 `xls:"VB每日签到"`
	WeeklySign    int64 `xls:"VB每周签到"`
	Recharge      int64 `xls:"累积充值金额"`
	DailyReward   int64 `xls:"每日奖励"`
	WeeklyReward  int64 `xls:"每周奖励"`
	UpgradeRewrad int64 `xls:"升级奖励"`
	WithdrawCount int   `xls:"每日提现笔数"`     // 每日提现次数
	Commission    int   `xls:"提现手续费（万分比）"` // 佣金比例
	VoiceSwitch   int   `xls:"麦克风功能"`      // 语音开关
	EmojiPrice    int   `xls:"互动表情价格"`     // 表情价格
	UnlockPhoto   int   `xls:"解锁头像"`       // 解锁头像
}

// 提现排行榜
type RankWithDraw struct {
	Id           string `xls:"人机Id"`
	RobotType    int32  `xls:"类型（0固定1随机）"`
	RankMinute   int32  `xls:"每X分钟触发提现判断"`
	WithdrawRate int32  `xls:"提现概率（万分比）"`
	Choices      ints   `xls:"提现档位范围"`
	VipChoices   ints   `xls:"人机VIP范围"`
	VipWeights   ints   `xls:"人机VIP权重"`
}

// 商城配置
type ShopConfig struct {
	NewbieGive       int64  `xls:"新客赠送"`
	RechargeInterval int64s `xls:"最低/最高充值金额"`
	WithdrawInterval int64s `xls:"最低/最高提现金额"`
	DefaultRecharge  int64s `xls:"默认充值商品金额"`
	VB               int64s `xls:"VB赠送"`
	DefaultWithdraw  int64s `xls:"默认提现商品金额"`
}

// 人机VIP配置
type VipRobot struct {
	Id         string `xls:"ID"`
	VipChoices ints   `xls:"人机VIP范围"`
	VipWeights ints   `xls:"人机VIP权重"`
}

// 小米14活动奖励配置表
type LuckyDrawReword struct {
	Level  int   `xls:"奖励等级"`
	Reward int64 `xls:"奖励彩金"`
	Limit  int32 `xls:"中奖人数"`
	Rate   int32 `xls:"每张号码中奖率（万分比）"`
}

// 小米14活动自动放号配置表
type LuckyDrawRobot struct {
	Id          string `xls:"ID"`
	Counts      int32s `xls:"已放号数量"`
	TimePeriod1 int32s `xls:"2:00:00-6:59:59"`
	TimePeriod2 int32s `xls:"7:00:00-12:59:59"`
	TimePeriod3 int32s `xls:"13:00:00-18:59:59"`
	TimePeriod4 int32s `xls:"19:00:00-1:59:59"`
}

// 大富翁基础配置
type ScratchTicketBase struct {
	Id     int `xls:"ID"`
	NextId int `xls:"下一级"`
	Cost   int `xls:"消耗bonus"`
	Cell   int `xls:"格子数"`
}

// 大富翁中奖配置
type ScratchTicketReward struct {
	Jackpot []int64 `xls:"奖金（1，2，3，4，5，6）"`
	Scope   []int64 `xls:"中奖范围"`
	Rule    []int   `xls:"对应配置"`
}

// 大富翁规则配置
type ScratchTicketRule struct {
	Id        int  `xls:"ID"`
	WeightPro ints `xls:"中奖概率"`
}

// 新提现跑马灯人机配置
type MarqueeWithdrawRobot struct {
	Id             int    `xls:"ID"`
	TimeRange      string `xls:"触发时间段"`
	Seconds        ints   `xls:"随机秒数"`
	Withdraws      ints   `xls:"提现档位范围"`
	WithdrawWeight ints   `xls:"提现档位权重"`
}

// 玩游戏抽奖
type PlayAndDrawActivity struct {
	Id           int     `xls:"ID"`                 // id
	ValidityTime int     `xls:"活动时效（小时）"`           // 有效期(小时)
	MinDrawTimes int     `xls:"领到奖励的抽奖次数"`          // 领到奖励的抽奖次数
	MaxPlayDraw  int     `xls:"玩游戏获得抽奖次数上限"`        // 玩游戏抽奖上限
	PlayRounds   ints    `xls:"玩N局游戏获得M次"`          // 玩N局游戏获得M次
	MaxShareDraw int     `xls:"分享获得抽奖次数上限"`         // 分享抽奖上限
	ShareFriends ints    `xls:"邀请N个玩家获得M次"`         // 邀请N个玩家获得M次
	Reward       int64s  `xls:"奖励金额（彩金，可提现，bonus）"` // 奖励
	RandSpan     float64 `xls:"随机跨度"`               // 随机跨度
	JumpRate     float64 `xls:"跃迁倍率"`               // 跃迁倍率
	DownRate     float64 `xls:"递减率"`                // 递减率
}

// 充值金额分类
type ChargeClassify struct {
	None   int64s `xls:"零充"`
	Normal int64s `xls:"普r"`
	XR     int64s `xls:"小r"`
	ZR     int64s `xls:"中r"`
	DR     int64s `xls:"大r"`
	CDR    int64s `xls:"超大r"`
}

func parseShare(f *excelize.File) (ret []Share, addr []ShareAddr, err error) {
	sheet := "t_share"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	sheet = "t_share_addr"
	err = decxls.UnmarshalExcelize(f, sheet, &addr)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseBeginner(f *excelize.File) (ret []Beginner, err error) {
	sheet := "t_beginner"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseRoomPeople(f *excelize.File) (ret []RoomPeople, err error) {
	sheet := "t_room_people"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseWeekCard(f *excelize.File) (ret []WeekCard, err error) {
	sheet := "t_week_card"
	err = decxls.UnmarshalExcelize(f, sheet, &ret)
	if err != nil {
		return
	}
	if len(ret) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet)
		return
	}
	return
}

func parseEmoji(f *excelize.File) (emoji []Emoji, emoji_config []EmojiConfig, err error) {
	sheet1 := "1.base"
	err = decxls.UnmarshalExcelize(f, sheet1, &emoji)
	if err != nil {
		return
	}
	if len(emoji) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}

	sheet2 := "2.config"
	err = decxls.UnmarshalExcelize(f, sheet2, &emoji_config)
	if err != nil {
		return
	}
	if len(emoji_config) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet2)
		return
	}
	return
}

func parseDailySign(f *excelize.File) (sign []DailySignIn, err error) {
	sheet1 := "t_sign"
	err = decxls.UnmarshalExcelize(f, sheet1, &sign)
	if err != nil {
		return
	}
	if len(sign) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}

	return
}

func parseOnlineReward(f *excelize.File) (online []OnlineReward, err error) {
	sheet1 := "t_online_reward"
	err = decxls.UnmarshalExcelize(f, sheet1, &online)
	if err != nil {
		return
	}
	if len(online) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}

	return
}

func parseLimitedGift(f *excelize.File) (gitf []LimitedGift, err error) {
	sheet1 := "t_limited_gift"
	err = decxls.UnmarshalExcelize(f, sheet1, &gitf)
	if err != nil {
		return
	}
	if len(gitf) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parsePVPRoom(f *excelize.File) (pvp []PvpRoom, err error) {
	sheet1 := "t_pvp_room"
	err = decxls.UnmarshalExcelize(f, sheet1, &pvp)
	if err != nil {
		return
	}
	if len(pvp) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseVip(f *excelize.File) (vip []VIP, err error) {
	sheet1 := "t_vip"
	err = decxls.UnmarshalExcelize(f, sheet1, &vip)
	if err != nil {
		return
	}
	if len(vip) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseRankWithdraw(f *excelize.File) (rank []RankWithDraw, err error) {
	sheet1 := "t_rank_withdraw"
	err = decxls.UnmarshalExcelize(f, sheet1, &rank)
	if err != nil {
		return
	}
	if len(rank) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseShop(f *excelize.File) (shop []ShopConfig, err error) {
	sheet1 := "t_shop"
	err = decxls.UnmarshalExcelize(f, sheet1, &shop)
	if err != nil {
		return
	}
	if len(shop) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseVipRobot(f *excelize.File) (vipRobots []VipRobot, err error) {
	sheet1 := "t_vip_robot"
	err = decxls.UnmarshalExcelize(f, sheet1, &vipRobots)
	if err != nil {
		return
	}
	if len(vipRobots) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseLuckyDraw(f *excelize.File) (rewords []LuckyDrawReword, robots []LuckyDrawRobot, err error) {
	sheet1, sheet2 := "t_lucky_draw_reward", "t_lucky_draw_robot"
	err = decxls.UnmarshalExcelize(f, sheet1, &rewords)
	if err != nil {
		return
	}
	if len(rewords) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	err = decxls.UnmarshalExcelize(f, sheet2, &robots)
	if err != nil {
		return
	}
	if len(rewords) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet2)
		return
	}
	return
}

func parseScratchTicket(f *excelize.File) (base []ScratchTicketBase, reward []ScratchTicketReward, rule []ScratchTicketRule, err error) {
	sheet1, sheet2, sheet3 := "t_scratch_ticket", "t_scratch_reward", "t_scratch_rule"
	err = decxls.UnmarshalExcelize(f, sheet1, &base)
	if err != nil {
		return
	}
	if len(base) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	err = decxls.UnmarshalExcelize(f, sheet2, &reward)
	if err != nil {
		return
	}
	if len(reward) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet2)
		return
	}
	err = decxls.UnmarshalExcelize(f, sheet3, &rule)
	if err != nil {
		return
	}
	if len(rule) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet3)
		return
	}
	return
}

func parseMarqueeWithdraw(f *excelize.File) (robots []MarqueeWithdrawRobot, err error) {
	sheet1 := "t_marquee_withdraw"
	err = decxls.UnmarshalExcelize(f, sheet1, &robots)
	if err != nil {
		return
	}
	if len(robots) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parsePlayAndShare(f *excelize.File) (acts []PlayAndDrawActivity, err error) {
	sheet1 := "t_play_draw"
	err = decxls.UnmarshalExcelize(f, sheet1, &acts)
	if err != nil {
		return
	}
	if len(acts) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func parseChargeClassify(f *excelize.File) (acts []ChargeClassify, err error) {
	sheet1 := "t_charge_classify"
	err = decxls.UnmarshalExcelize(f, sheet1, &acts)
	if err != nil {
		return
	}
	if len(acts) == 0 {
		err = fmt.Errorf("配置表解析错误: %s", sheet1)
		return
	}
	return
}

func updateHall(d []byte, save bool, filename string) (err error) {
	byte_reader := bytes.NewReader(d)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		return err
	}

	switch filename {
	case "分享":
		updateShare(save, f)
	case "新手模式":
		updateBeginner(save, f)
	case "房间人数":
		updateRoomPeople(save, f)
	case "周卡":
		updateWeekCard(save, f)
	case "人机表情":
		updateEmoji(save, f)
	case "每日签到":
		updateDailySign(save, f)
	case "在线奖励":
		updateOnlineReward(save, f)
	case "限时礼包":
		updateLimitedGift(save, f)
	case "对战房基本配置表":
		err = updatePvpRoom(save, f)
	case "vip":
		err = updateVIP(save, f)
	case "提现排行榜":
		err = updateRankWithdraw(save, f)
	case "商城配置":
		err = updateShopConfig(save, f)
	case "人机VIP等级配置":
		err = updateVipRobot(save, f)
	case "抽小米手机活动":
		err = updateLuckyDraw(save, f)
	case "大富翁":
		err = updateScratchTicket(save, f)
	case "新跑马灯配置":
		err = updateMarqueeWithdraw(save, f)
	case "游戏分享抽奖":
		err = updatePlayAndShare(save, f)
	case "crash策略配置表":
		err = updateCrashStrategy(save, f)
	case "充值金额分类":
		err = updateChargeClassify(save, f)
	case "lhd策略配置表":
		err = updateLHDStrategy(save, f)
	case "7up策略配置表":
		err = updateUPStrategy(save, f)
	case "ab策略配置表":
		err = updateABStrategy(save, f)
	case "l3策略配置表":
		err = updateCPStrategy(save, f)
	case "rb策略配置表":
		err = updateRBStrategy(save, f)
	case "aviator策略配置表":
		err = updateAvStrategy(save, f)
	default:
		return errors.New("filename is nil")
	}

	return
}

func updateShare(save bool, f *excelize.File) error {
	shares, addrs, err1 := parseShare(f)
	if err1 != nil {
		return err1
	}

	var datas []data.Share
	for _, s := range shares {
		data := data.Share{
			Id:            s.Id,
			Recharge:      s.Recharge,
			FirstReward:   s.FirstReward,
			SecondReward:  s.SecondReward,
			FirstRecharge: s.FirstRecharge,
			MaxPeople:     s.MaxPeople,
			Url:           s.Url,
			PCUrl:         s.PCUrl,
		}
		datas = append(datas, data)
	}

	for _, s := range datas {
		if save {
			s.Save()
		}
		config.SetShare(s)
	}

	var addrdatas []data.ShareAddr
	for _, s := range addrs {
		data := data.ShareAddr{
			Id:   s.Id,
			Link: s.Link,
		}
		addrdatas = append(addrdatas, data)
	}

	for _, s := range addrdatas {
		if save {
			s.Save()
		}
		config.SetShareAddr(s)
	}
	return nil
}

func updateBeginner(save bool, f *excelize.File) error {
	beginners, err1 := parseBeginner(f)
	if err1 != nil {
		return err1
	}

	var datas []data.Beginner
	for _, s := range beginners {
		data := data.Beginner{
			Id:               s.Id,
			RegistMode1:      s.RegistMode1.Value,
			RegistMode2:      s.RegistMode2.Value,
			NoviceGame:       s.NoviceGame.Value,
			Recharge:         s.Recharge,
			PlayerGames:      s.PlayerGames,
			OnlineTime:       s.OnlineTime,
			OutCashInterval:  s.OutCashInterval.Value,
			Winning:          s.Winning.Value,
			OutCashLimited:   s.OutCashLimited,
			FrothMaxRecharge: s.FrothMaxRecharge,
			FrothGiftRate:    s.FrothGiftRate,
			FrothWonRate:     s.FrothWonRate.Value,
			FrothBetInterval: s.FrothBetInterval.Value,
			FactorSeed:       s.FactorSeed,
			NewbiewToCivil:   s.NewbiewToCivil.Value,
		}
		datas = append(datas, data)
	}

	for _, s := range datas {
		if save {
			s.Save()
		}
		config.SetBeginner(s)
	}
	return nil
}

func updateRoomPeople(save bool, f *excelize.File) error {
	peoples, err1 := parseRoomPeople(f)
	if err1 != nil {
		return err1
	}
	var datas []data.RoomPeople
	for _, s := range peoples {
		data := data.RoomPeople{
			ID:              s.ID,
			Gtype:           s.Gtype,
			RoomId:          s.RoomId,
			PeopleInterval1: s.PeopleInterval1.Value,
			PeopleInterval2: s.PeopleInterval2.Value,
			PeopleInterval3: s.PeopleInterval3.Value,
		}
		datas = append(datas, data)
	}

	for _, s := range datas {
		if save {
			s.Save()
		}
		config.SetRoomPeople(s)
	}
	return nil
}

func updateWeekCard(save bool, f *excelize.File) error {
	cards, err1 := parseWeekCard(f)
	if err1 != nil {
		return err1
	}
	var datas []data.WeeklyCard
	for _, s := range cards {
		data := data.WeeklyCard{
			ID:     s.ID,
			Price:  s.Price,
			Reward: s.Reward,
			Give:   s.Give.Value,
		}
		datas = append(datas, data)
	}

	for _, s := range datas {
		if save {
			s.Save()
		}
		config.SetWeeklyCard(s)
	}
	return nil
}

func updateEmoji(save bool, f *excelize.File) error {
	emoji, emoji_config, err := parseEmoji(f)
	if err != nil {
		return err
	}

	var configs []data.EmojiConfig
	for _, v := range emoji_config {
		config := data.EmojiConfig{
			Id:                 v.Id,
			DeskSendProb:       v.DeskSendProb,
			DeskSendWeight:     v.DeskSendWeight.Value,
			DeskSendDelay:      v.DeskSendDelay.Value,
			GameStartProb:      v.GameStartProb,
			GameStartWeight:    v.GameStartWeight.Value,
			GameStartDelay:     v.GameStartDelay.Value,
			WaitLongProb:       v.WaitLongProb,
			WaitLongWeight:     v.WaitLongWeight.Value,
			WaitLongDelay:      v.WaitLongDelay.Value,
			ReplyProb:          v.ReplyProb,
			ReplyWeight:        v.ReplyWeight.Value,
			ReplyDelay:         v.ReplyDelay.Value,
			OtherSeeProb:       v.OtherSeeProb,
			OtherSeeWeight:     v.OtherSeeWeight.Value,
			OtherSeeDelay:      v.OtherSeeDelay.Value,
			OtherRaiseProb:     v.OtherRaiseProb,
			OtherRaiseWeight:   v.OtherRaiseWeight.Value,
			OtherRaiseDelay:    v.OtherRaiseDelay.Value,
			OtherBiProb:        v.OtherBiProb.Value,
			OtherBiWeight:      v.OtherBiWeight.Value,
			OtherBiDelay:       v.OtherBiDelay.Value,
			SelfBiWinProb:      v.SelfBiWinProb,
			SelfBiWinWeight:    v.SelfBiWinWeight.Value,
			SelfBiWinDelay:     v.SelfBiWinDelay.Value,
			MoPreCardProb:      v.MoPreCardProb,
			MoPreCardWeight:    v.MoPreCardWeight.Value,
			MoPreCardDelay:     v.MoPreCardDelay.Value,
			OtherChuCardProb:   v.OtherChuCardProb,
			OtherChuCardWeight: v.OtherChuCardWeight.Value,
			OtherChuCardDelay:  v.OtherChuCardDelay.Value,
		}
		configs = append(configs, config)
	}

	var emojis []data.Emoji
	for _, v := range emoji {
		e := data.Emoji{
			Id:              v.Id,
			RobotTypeWeight: v.RobotTypeWeight.Value,
			RobotType:       v.RobotType.Value,
			EmojiConfig:     configs,
		}
		emojis = append(emojis, e)
	}

	for _, v := range emojis {
		if save {
			v.Save()
		}
		config.SetEmoji(v)
	}
	return nil
}

func updateDailySign(save bool, f *excelize.File) error {
	signs, err := parseDailySign(f)
	if err != nil {
		return err
	}

	var configs []data.DailySign
	for _, v := range signs {
		config := data.DailySign{
			Id:     v.ID,
			Number: v.Reward,
			Price:  v.Recharge,
		}
		configs = append(configs, config)
	}

	for _, v := range configs {
		if save {
			v.Save()
		}
		config.SetDailySign(v)
	}
	return nil
}

func updateOnlineReward(save bool, f *excelize.File) error {
	onlines, err := parseOnlineReward(f)
	if err != nil {
		return err
	}

	var configs []data.OnlineReward
	for _, v := range onlines {
		config := data.OnlineReward{
			Id:         v.ID,
			OnlineTime: int64(v.OnlineTime),
			Reward:     v.Reward.Value,
		}
		configs = append(configs, config)
	}

	for _, v := range configs {
		if save {
			v.Save()
		}
		config.SetOnlineReward(v)
	}
	return nil
}

func updateLimitedGift(save bool, f *excelize.File) error {
	gifts, err := parseLimitedGift(f)
	if err != nil {
		return err
	}

	var configs []data.LimitedGift
	for _, v := range gifts {
		config := data.LimitedGift{
			Id:         v.ID,
			Switch:     v.Switch,
			Price:      v.Price,
			Reward:     v.Reward,
			DuringTime: v.DuringTime,
			Recharge:   v.Recharge,
			GameRound:  v.GameRound,
		}
		configs = append(configs, config)
	}

	for _, v := range configs {
		if save {
			v.Save()
		}
		config.SetLimitedGift(v)
	}
	return nil
}

func updatePvpRoom(save bool, f *excelize.File) error {
	pvps, err := parsePVPRoom(f)
	if err != nil {
		return err
	}

	v := pvps[0]
	pvpRoom := &data.PvpRoom{
		Id:             1,
		PvpChargeLimit: v.PvpChargeLimit,
		GameStartWait:  v.GameStartWait,
		TpFunInitScore: v.TpFunInitScore,
		RmFunInitScore: v.RmFunInitScore,
		AbFunInitScore: v.AbFunInitScore,
		TpGoldRound:    v.TpGoldRound.Value,
		RmGoldRound:    v.RmGoldRound.Value,
		AbGoldRound:    v.AbGoldRound.Value,
		TpFunRoundCost: v.TpFunRoundCost.Value,
		RmFunRoundCost: v.RmFunRoundCost.Value,
		AbFunRoundCost: v.AbFunRoundCost.Value,
	}
	if save {
		pvpRoom.Save()
	}
	config.SetPvpRoom(pvpRoom)
	return nil
}

func updateVIP(save bool, f *excelize.File) error {
	vips, err := parseVip(f)
	if err != nil {
		return err
	}

	for _, v := range vips {
		vip := &data.Vip{
			Id:         v.Level,
			Recharge:   v.Recharge,
			NextLevel:  v.NextLevel,
			DailySign:  v.DailySign,
			WeeklySign: v.WeeklySign,
			// DailyReward:   v.DailyReward,
			// WeeklyReward:  v.WeeklyReward,
			// UpgradeRewrad: v.UpgradeRewrad,
			WithdrawCount: v.WithdrawCount,
			Commission:    v.Commission,
			EmojiPrice:    v.EmojiPrice,
			VoiceSwitch:   v.VoiceSwitch,
			UnlockPhoto:   v.UnlockPhoto,
		}
		if save {
			vip.Save()
		}
		config.SetVip(*vip)
	}

	return nil
}

func updateRankWithdraw(save bool, f *excelize.File) error {
	ranks, err := parseRankWithdraw(f)
	if err != nil {
		return err
	}

	vals := make([]*data.RankWithdraw, 0, len(ranks))
	for _, v := range ranks {
		rank := &data.RankWithdraw{
			Id:           v.Id,
			RobotType:    v.RobotType,
			RankMinute:   v.RankMinute,
			WithdrawRate: v.WithdrawRate,
			Choices:      v.Choices.Value,
			VipChoices:   v.VipChoices.Value,
			VipWeights:   v.VipWeights.Value,
		}
		if save {
			rank.Save()
		}
		vals = append(vals, rank)
	}
	config.SetRankWithdraw(vals)
	return nil
}

func updateShopConfig(save bool, f *excelize.File) error {
	shops, err := parseShop(f)
	if err != nil {
		return err
	}

	shopDatas := make([]data.Shop, 0, len(shops))
	shop := data.Shop{
		Id:               "1",
		NewbieGive:       shops[0].NewbieGive,
		RechargeInterval: shops[0].RechargeInterval.Value,
		WithdrawInterval: shops[0].WithdrawInterval.Value,
		DefaultRecharge:  shops[0].DefaultRecharge.Value,
		DefaultWithdraw:  shops[0].DefaultWithdraw.Value,
		VB:               shops[0].VB.Value,
	}
	shopDatas = append(shopDatas, shop)

	if save {
		data.ClearShop()
	}

	for _, shop := range shopDatas {
		if save {
			shop.Save()
		}
		config.SetShop(shop)
	}
	return nil
}

func updateVipRobot(save bool, f *excelize.File) error {
	vipRobots, err := parseVipRobot(f)
	if err != nil {
		return err
	}

	vals := make([]*data.VipRobot, 0, len(vipRobots))
	for _, v := range vipRobots {
		vr := &data.VipRobot{
			Id:         v.Id,
			VipChoices: v.VipChoices.Value,
			VipWeights: v.VipWeights.Value,
		}
		if save {
			vr.Save()
		}
		vals = append(vals, vr)
	}
	config.SetVipRobot(vals)
	return nil
}

func updateLuckyDraw(save bool, f *excelize.File) error {
	rewards, robots, err := parseLuckyDraw(f)
	if err != nil {
		return err
	}
	vals1 := make([]*data.LuckyDrawReward, 0, len(rewards))
	for _, v := range rewards {
		reword := &data.LuckyDrawReward{
			Id:     v.Level,
			Reward: v.Reward,
			Limit:  v.Limit,
			Rate:   v.Rate,
		}
		if save {
			reword.Save()
		}
		vals1 = append(vals1, reword)
	}
	config.SetLuckyDrawRewards(vals1)

	vals2 := make([]*data.LuckyDrawRobot, 0, len(rewards))
	for _, v := range robots {
		robot := &data.LuckyDrawRobot{
			Id:          v.Id,
			Counts:      v.Counts.Value,
			TimePeriod1: v.TimePeriod1.Value,
			TimePeriod2: v.TimePeriod2.Value,
			TimePeriod3: v.TimePeriod3.Value,
			TimePeriod4: v.TimePeriod4.Value,
		}
		if save {
			robot.Save()
		}
		vals2 = append(vals2, robot)
	}
	config.SetLuckyDrawRobots(vals2)
	return nil
}

func updateScratchTicket(save bool, f *excelize.File) error {
	bases, rewards, rules, err := parseScratchTicket(f)
	if err != nil {
		return err
	}
	vals1 := make([]*data.ScratchTicketConfig, 0, len(bases))
	for _, v := range bases {
		base := &data.ScratchTicketConfig{
			Id:     v.Id,
			NextId: v.NextId,
			Cost:   int64(v.Cost),
			Cell:   v.Cell,
		}

		base.Jackpot = rewards[0].Jackpot
		base.Scope = rewards[0].Scope
		base.Rule = rewards[0].Rule
		for _, r := range rules {
			s := data.ScratchTicketRule{
				Id:        r.Id,
				WeightPro: r.WeightPro.Value,
			}
			base.RuleConfig = append(base.RuleConfig, s)
		}

		if save {
			base.Save()
		}
		vals1 = append(vals1, base)
	}
	config.SetScratchTicketConfig(vals1)

	return nil
}

func updateMarqueeWithdraw(save bool, f *excelize.File) error {
	robots, err := parseMarqueeWithdraw(f)
	if err != nil {
		return err
	}

	vals2 := make([]*data.MarqueeWithdraw, 0, len(robots))
	for _, v := range robots {
		robot := &data.MarqueeWithdraw{
			Id:             v.Id,
			TimeRange:      v.TimeRange,
			Seconds:        v.Seconds.Value,
			Withdraws:      v.Withdraws.Value,
			WithdrawWeight: v.WithdrawWeight.Value,
		}
		if save {
			robot.Save()
		}
		vals2 = append(vals2, robot)
	}
	config.SetMarqueeWithdraws(vals2)
	return nil
}

func updatePlayAndShare(save bool, f *excelize.File) error {
	// acts, err := parsePlayAndShare(f)
	// if err != nil {
	// 	return err
	// }

	// for _, v := range acts {
	// 	act := data.PlayAndDrawActivity{
	// 		Id:           v.Id,
	// 		ValidityTime: v.ValidityTime,
	// 		MinDrawTimes: v.MinDrawTimes,
	// 		MaxPlayDraw:  v.MaxPlayDraw,
	// 		MaxShareDraw: v.MaxShareDraw,
	// 		Reward:       v.Reward.Value,
	// 		RandSpan:     v.RandSpan,
	// 		JumpRate:     v.JumpRate,
	// 		DownRate:     v.DownRate,
	// 		ShareFriends: v.ShareFriends.Value,
	// 		PlayRounds:   v.PlayRounds.Value,
	// 	}
	// 	if save {
	// 		act.Save()
	// 	}
	// 	config.SetPlayShareConfig(act)
	// }
	return nil
}

func updateChargeClassify(save bool, f *excelize.File) error {
	acts, err := parseChargeClassify(f)
	if err != nil {
		return err
	}

	for _, v := range acts {
		act := data.ChargeClassify{
			Id:     1,
			None:   v.None.Value,
			Normal: v.Normal.Value,
			XR:     v.XR.Value,
			ZR:     v.ZR.Value,
			DR:     v.DR.Value,
			CDR:    v.CDR.Value,
		}
		if save {
			act.Save()
		}
		config.SetChargeClassify(act)
	}
	return nil
}
