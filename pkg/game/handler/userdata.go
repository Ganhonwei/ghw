package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/utils"
	"time"
)

var (
	pkWallet = map[string]struct{}{
		"JAZZCASH":  {},
		"EASYPAISA": {},
	}
)

// GetCurrency 获取货币信息
func GetCurrency(ctos *pb.GetCurrencyReq, p *data.User) (stoc *pb.GetCurrencyRsp) {
	stoc = new(pb.GetCurrencyRsp)
	if p == nil {
		return
	}
	stoc.Data = PackCurrency(p)
	return
}

// PackCurrency 打包基础货币消息
func PackCurrency(p *data.User) (msg *pb.Currency) {
	return &pb.Currency{
		Coin:    p.GetCoin(),
		Diamond: p.GetDiamond(),
	}
}

// CurrencyMsg 变动货币消息
func CurrencyMsg(diamond, coin, card, vb, outDiamond int64) (msg *pb.Currency) {
	return &pb.Currency{
		Diamond:    diamond,
		Coin:       coin,
		Card:       card,
		Vb:         vb,
		OutDiamond: outDiamond,
	}
}

// PushCurrencyMsg 货币变动推送消息
func PushCurrencyMsg(diamond, coin, card, chip, outDiamond int64,
	ltype int32) (msg *pb.PushCurrencyNtf) {
	msg = new(pb.PushCurrencyNtf)
	msg.Rtype = uint32(ltype)
	msg.Data = CurrencyMsg(diamond, coin, card, chip, outDiamond)
	msg.Vbtype = GetBonusChangeType(ltype, chip)
	return
}

// ChangeCurrencyMsg 货币变动消息
func ChangeCurrencyMsg(diamond, coin, card, chip, give, out int64,
	ltype int32, userid, desc, waterid string) (msg *pb.ChangeCurrency) {
	msg = ChangeCurrencyMsg1(diamond, coin,
		card, chip, give, out, ltype, userid, desc, waterid, false)
	return
}

func ChangeCurrencyMsg1(diamond, coin, card, chip, give, out int64,
	ltype int32, userid, desc, waterid string, control bool) (msg *pb.ChangeCurrency) {
	msg = new(pb.ChangeCurrency)
	msg.Type = ltype
	msg.Diamond = diamond
	msg.Coin = coin
	msg.Card = card
	msg.Chip = chip
	msg.Userid = userid
	msg.Desc = desc
	msg.WaterId = waterid
	msg.Control = control
	msg.Give = give
	msg.Out = out
	return
}

// Pay2ChangeCurr 货币变动消息
func Pay2ChangeCurr(arg *pb.PayCurrency) (msg *pb.ChangeCurrency) {
	msg = &pb.ChangeCurrency{
		Userid:  arg.Userid,
		Type:    arg.Type,
		Coin:    arg.Coin,
		Diamond: arg.Diamond,
		Give:    arg.Give,
		Chip:    arg.Chip,
		Card:    arg.Card,
		Desc:    arg.Desc,
		WaterId: "",
	}
	return
}

// 货币修改变得消息
func ModifyChangeCurr(user *data.User, arg *pb.ModifyCurrency) (msg *pb.ChangeCurrency) {
	var realDiamond int64 = 0
	var realGive int64 = 0
	if arg.Diamond >= 0 {
		realDiamond = arg.Diamond - user.Diamond
	}
	if arg.Give >= 0 {
		realGive = arg.Give - user.VBBank
	}
	msg = &pb.ChangeCurrency{
		Userid:  arg.Userid,
		Type:    arg.Type,
		Diamond: realDiamond,
		Give:    realGive,
		Desc:    "后台修改货币",
		WaterId: "",
	}
	return
}

// Offline2Change 货币变动消息
func Offline2Change(arg *pb.OfflineCurrency) (msg *pb.ChangeCurrency) {
	msg = &pb.ChangeCurrency{
		Userid:  arg.Userid,
		Type:    arg.Type,
		Coin:    arg.Coin,
		Diamond: arg.Diamond,
		Chip:    arg.Chip,
		Card:    arg.Card,
		Desc:    arg.Desc,
		WaterId: arg.WaterId,
		Control: arg.Control,
		Give:    arg.Give,
		Out:     arg.Out,
	}
	return
}

// Ping 心跳请求
func Ping(ctos *pb.PingReq) (stoc *pb.PingRsp) {
	stoc = new(pb.PingRsp)
	stoc.Ctime = ctos.Ctime
	stoc.Stime = uint32(time.Now().Unix())
	// stoc.Time = ctos.GetTime()
	return
}

// GetUserDataMsg 获取自己数据
func GetUserDataMsg(ctos *pb.UserDataReq, p *data.User) (stoc *pb.UserDataRsp) {
	stoc = new(pb.UserDataRsp)
	stoc.Data = new(pb.UserData)
	userid := ctos.GetUserid()
	if userid == "" {
		stoc.Error = pb.UsernameEmpty
		return
	}
	stoc.Data = PackUserData(p)
	stoc.Info = PackUserTop(p)
	return
}

// PackUserData 打包基础数据
func PackUserData(p *data.User) (stoc *pb.UserData) {
	guids := make([]*pb.GameGuid, 0)
	for _, gg := range p.GameGuildSlice {
		guids = append(guids, gg.BuildGuid())
	}
	return &pb.UserData{
		Userid:            p.GetUserid(),
		Token:             p.GetToken(),
		Nickname:          p.GetNickname(),
		Phone:             p.GetPhone(),
		Sex:               p.GetSex(),
		Photo:             p.Photo,
		Coin:              p.GetCoin(),
		Diamond:           p.GetDiamond(),
		OutDiamond:        p.OutDiamond,
		Sign:              p.GetSign(),
		Recharge:          p.FirstRecharge != "",
		RechargeActivity:  p.NoviceGift != "",
		OnlineReward:      p.OnlineReward,
		Phone2:            p.Phone2,
		BankName:          p.Bank,
		BankNumber:        p.BankAccounts,
		BankAccountHolder: p.BankAccountHolder,
		Email:             p.Email,
		Ifsc:              p.IFSC,
		Usdt:              p.USDT,
		RegistMode:        p.RegistMode,
		RegistReward:      p.RegistReward,
		Reward:            BuildRegistReward(p.RegistArea),
		FirstWithdraw:     p.WithdrawWindow,
		FirstWithdraw2:    p.WithdrawWindow2,
		BindPhone:         p.FirstLoginToday && p.RegistMode == 1 && p.LoginTimes == 2,
		FirstLogin:        p.FirstLoginToday,
		Guids:             guids,
		Money:             int64(p.Money),
		UserType:          int32(p.RegistArea),
		Vb:                p.VBBank,
		VbCash:            p.UnlockBonus,
		VbCashout:         p.CashOutBonus,
		PkBank:            BuildPKBankInfo(p),
	}
}

// PackUserTop 玩家个人数据
func PackUserTop(p *data.User) (stoc *pb.TopInfo) {
	return &pb.TopInfo{
		//Topcoins:      p.GetTopCoins(),      //最高拥有金币总金额
		//Topdiamonds:   p.GetTopDiamonds(),   //最高拥有钻石总金额
		//Topcards:      p.GetTopCards(),      //最高拥有房卡总数
		// Topchips: p.GetTopChips(), //最高拥有筹码总金额
		//Topwincoin:    p.GetTopWinCoin(),    //单局赢最高金币金额
		//Topwindiamond: p.GetTopWinDiamond(), //单局赢最高钻石金额
		// Topwinchip: p.GetTopWinChip(),                              //单局赢最高筹码金额
		Registtime: utils.Format("Y-m-d H:i:s", p.GetRegistTime()), //加入游戏时间
		Logintime:  utils.Format("Y-m-d H:i:s", p.GetLoginTime()),  //最后登录时间
	}
}

// GetUserData 获取其它玩家数据
func GetUserData(p *data.User) (stoc *pb.GotUserData) {
	stoc = new(pb.GotUserData)
	if p == nil {
		stoc.Error = pb.UsernameEmpty
		return
	}
	//基本数据
	stoc.Data = PackUserData(p)
	stoc.Info = PackUserTop(p)
	return
}

// UserDataMsg 获取其它玩家数据消息
func UserDataMsg(p *pb.GotUserData) (stoc *pb.UserDataRsp) {
	stoc = new(pb.UserDataRsp)
	if p == nil {
		stoc.Error = pb.UsernameEmpty
		return
	}
	if p.Error != pb.OK {
		stoc.Error = p.Error
		return
	}
	//基本数据
	stoc.Data = p.GetData()
	stoc.Info = p.GetInfo()
	return
}

// PackRankMsg 获取排行榜信息
// func PackRankMsg() (msg *pb.RankRsp) {
// 	msg = new(pb.RankRsp)
// 	list, err := data.GetRank()
// 	if err != nil {
// 		glog.Errorf("GetRank err %v", err)
// 	}
// 	glog.Debugf("rank list %#v", list)
// 	for _, v := range list {
// 		msg2 := new(pb.Rank)
// 		if val, ok := v["coin"]; ok {
// 			msg2.Coin = val.(int64)
// 		}
// 		if val, ok := v["_id"]; ok {
// 			msg2.Userid = val.(string)
// 		}
// 		if val, ok := v["nickname"]; ok {
// 			msg2.Nickname = val.(string)
// 		}
// 		if val, ok := v["photo"]; ok {
// 			msg2.Photo = val.(string)
// 		}
// 		if val, ok := v["sign"]; ok {
// 			msg2.Sign = val.(string)
// 		}
// 		if val, ok := v["weixin"]; ok {
// 			msg2.Weixin = val.(string)
// 		}
// 		if msg2.Userid == "" {
// 			continue
// 		}
// 		msg.List = append(msg.List, msg2)
// 	}
// 	return
// }

// GiveBankMsg 银行变动消息
func GiveBankMsg(coin int64, ltype int32, userid, from string) (msg *pb.BankGive) {
	msg = new(pb.BankGive)
	msg.Type = ltype
	msg.Coin = coin
	msg.Userid = userid
	msg.From = from
	return
}

// BankChangeMsg 银行变动消息
func BankChangeMsg(coin int64, ltype int32, userid, from string) (msg *pb.BankChange) {
	msg = new(pb.BankChange)
	msg.Type = ltype
	msg.Coin = coin
	msg.Userid = userid
	msg.From = from
	return
}

// 修改玩家信息
func ModifyUserData(user *data.User, data *data.ModifyUserData) {
	user.Photo = data.Photo
	user.Nickname = data.NickName
	user.Phone2 = data.Phone
	user.Bank = data.BankName
	user.BankAccounts = data.BankNumber
	user.IFSC = data.Ifsc

	user.AD_BundleId = data.BundleId
	// user.AD_Tracker_Channel = data.MediaSource
	user.AD_ADID = data.AfId
	user.AD_Key = data.AfKey
	user.AD_OsVersion = data.OsVersion
	user.AD_AppId = data.AppId
	user.AD_RefGameId = data.RefGameId
	user.AD_RefPkgName = data.RefPkgName
}

func PointControl(user *data.User, data *data.PointControl) {
	user.PCSwitch = data.Switch
	user.PCFactor = data.Factor
	user.PCScore = data.Score
}

// 正常切换为泡沫
func NormalToFrothState(user *data.User) bool {
	if user.State != data.NormalState {
		return false
	}
	bean := config.GetBeginner(1)
	if bean.Id == 0 {
		return false
	}

	if user.Money >= uint32(bean.FrothMaxRecharge) {
		// 充值超过泡沫充值最大值
		return false
	}

	winScroe := user.Diamond + int64(user.CashOut) - int64(user.Money) // 当前赢分
	// 当前赢分/已充值金额 > 30%
	if user.Money > 0 && winScroe*10000/int64(user.Money) > int64(bean.FrothGiftRate) {
		// 进入泡沫状态
		return true
	}
	return false
}

// 泡沫切换为正常
func FrothToNormalState(user *data.User) bool {
	if user.State != data.FrothState {
		return false
	}
	if user.IsB() {
		return true
	}
	bean := config.GetBeginner(1)
	if bean.Id == 0 {
		return false
	}

	if user.Money >= uint32(bean.FrothMaxRecharge) {
		// 充值超过泡沫充值最大值
		return true
	}

	winScroe := user.Diamond + int64(user.CashOut) - int64(user.Money) // 当前赢分
	// 当前赢分/已充值金额 > 30%
	if user.Money > 0 && winScroe*10000/int64(user.Money) > int64(bean.FrothGiftRate) {
		// 进入泡沫状态
		return false
	}
	return true
}

// GetBonusChangeType bonus 变更类型
// 1.Recharging: 充值礼包获得bonus
// 2.Buy a Weekly Plan: 周卡
// 3.Balance Interest: 因为利息获得bonus
// 4.VIP Level Up: VIP升级 bonus
// 5.VIP Weekly Bonus: VIP每周返利
// 6.Trial Period Winnings And Bonuses: 转成付费玩家导致试玩期间的token转化为bonus时
// 7.Bonus Unlocked And Claimed: 当玩家因为解锁到足够金额而领走bonus
// 8.Ranking: 获得排行榜奖金
// 9.Prize Wheel: 转盘奖金
// 10.overtime compensation: 提现超时赔付金
// 11.refund compensation: 提现退款补偿金
// 12.agent rebate: 代理打码返佣
// 13.per-referral bonus: 推荐充值玩家人头奖励
// 14.referral milestone bonus: 推荐充值玩家达成里程碑获得bonus
// 101.Platform Giveway: 因为上述其他原因获得bonus
// 102.User Claimed: 因为上述其他原因减少bonus
// vb ltypes: 46, 57, 58, 59, 61, 62, 64, 74, 80, 82, 84, 86, 97, 98, 103, 119, 121, 123, 124, 126, 129, 132
func GetBonusChangeType(ltype int32, bonus int64) int32 {
	if bonus == 0 {
		return 0
	}

	if utils.SliceIn(ltype,
		int32(pb.LOG_TYPE57),
		int32(pb.LOG_TYPE74),
		int32(pb.LOG_TYPE80),
		int32(pb.LOG_TYPE80),
		int32(pb.LOG_TYPE84),
		int32(pb.LOG_TYPE86),
		int32(pb.LOG_TYPE86),
		int32(pb.LOG_TYPE129),
	) {
		return 1
	}

	switch ltype {
	case int32(pb.LOG_TYPE59): //VB利息增长
		return 2
	case int32(pb.LOG_TYPE121): //VB利息增长
		return 3
	case int32(pb.LOG_TYPE99): //vip等级奖励
		return 4
	case int32(pb.LOG_TYPE98): //vip每周
		return 5
	case int32(pb.LOG_TYPE103): //转为正常玩家
		return 6
	case int32(pb.LOG_TYPE132): //VB提取
		return 7
	}

	switch ltype {
	case int32(pb.LOG_TYPE134): //打码排行榜
		return 8
	case int32(pb.LOG_TYPE135): //转盘
		return 9
	case int32(pb.LOG_TYPE138): //提现超时赔付领取
		return 10
	case int32(pb.LOG_TYPE137): //用户主动提现申请退款补偿金
		return 11
	case int32(pb.LOG_TYPE142): //领取代理打码返佣金额
		return 12
	case int32(pb.LOG_TYPE141): //领取代理人头返佣金额
		return 13
	case int32(pb.LOG_TYPE144): //领取代理人头返佣金额
		return 14
	}

	if bonus > 0 {
		return 101
	} else {
		return 102
	}
}

// 根据bonus type反差 logtypes
func GetBonusChangeType2Ltypes(btype int32) (ltypes, notLtypes []int32, sub int8) {

	if btype == 1 {
		ltypes = append(ltypes,
			int32(pb.LOG_TYPE57),
			int32(pb.LOG_TYPE74),
			int32(pb.LOG_TYPE80),
			int32(pb.LOG_TYPE80),
			int32(pb.LOG_TYPE84),
			int32(pb.LOG_TYPE86),
			int32(pb.LOG_TYPE86),
			int32(pb.LOG_TYPE129))
		return
	}

	switch btype {
	case 2: //VB利息增长
		ltypes = append(ltypes, int32(pb.LOG_TYPE59))
		return
	case 3: //VB利息增长
		ltypes = append(ltypes, int32(pb.LOG_TYPE121))
		return
	case 4: //vip等级奖励
		ltypes = append(ltypes, int32(pb.LOG_TYPE99))
		return
	case 5: //vip每周
		ltypes = append(ltypes, int32(pb.LOG_TYPE98))
		return
	case 6: //转为正常玩家
		ltypes = append(ltypes, int32(pb.LOG_TYPE103))
		return
	case 7: //VB提取
		ltypes = append(ltypes, int32(pb.LOG_TYPE132))
		return
	}

	switch btype {
	case 8: //打码排行榜
		ltypes = append(ltypes, int32(pb.LOG_TYPE134))
		return
	case 9: //转盘
		ltypes = append(ltypes, int32(pb.LOG_TYPE135))
		return
	case 10: //提现超时赔付领取
		ltypes = append(ltypes, int32(pb.LOG_TYPE138))
		return
	case 11: //用户主动提现申请退款补偿金
		ltypes = append(ltypes, int32(pb.LOG_TYPE137))
		return
	case 12: //领取代理打码返佣金额
		ltypes = append(ltypes, int32(pb.LOG_TYPE142))
		return
	case 13: //领取代理人头返佣金额
		ltypes = append(ltypes, int32(pb.LOG_TYPE141))
		return
	case 14: //领取代理人头返佣金额
		ltypes = append(ltypes, int32(pb.LOG_TYPE144))
		return
	}

	if btype == 101 {
		sub = 1
	} else if btype == 102 {
		sub = 2
	} else {
		return
	}

	notLtypes = append(notLtypes,
		int32(pb.LOG_TYPE57),
		int32(pb.LOG_TYPE74),
		int32(pb.LOG_TYPE80),
		int32(pb.LOG_TYPE80),
		int32(pb.LOG_TYPE84),
		int32(pb.LOG_TYPE86),
		int32(pb.LOG_TYPE86),
		int32(pb.LOG_TYPE129))
	notLtypes = append(notLtypes,
		int32(pb.LOG_TYPE59),
		int32(pb.LOG_TYPE121),
		int32(pb.LOG_TYPE99),
		int32(pb.LOG_TYPE98),
		int32(pb.LOG_TYPE103),
		int32(pb.LOG_TYPE132),
		int32(pb.LOG_TYPE134),
		int32(pb.LOG_TYPE135),
		int32(pb.LOG_TYPE138),
		int32(pb.LOG_TYPE137),
		int32(pb.LOG_TYPE142),
		int32(pb.LOG_TYPE141),
		int32(pb.LOG_TYPE144))
	return
}
