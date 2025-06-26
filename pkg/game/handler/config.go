package handler

import (
	"fmt"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"

	jsoniter "github.com/json-iterator/go"
)

//单条配置修改同步都按map格式

// SyncConfig 同步配置
func SyncConfig(msg *pb.SyncConfig, node string) (err error) {
	switch msg.Type {
	case pb.CONFIG_ENV: //变量
		b := make(map[string]int32)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelEnv(k)
			case pb.CONFIG_UPSERT:
				config.SetEnv(k, v)
			}
		}
	case pb.CONFIG_NOTICE: //公告
		b := make(map[string]data.Notice)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}

		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelNotice(k)
			case pb.CONFIG_UPSERT:
				//玩家个人消息直接写数据库，不缓存在内存中
				// if v.Userid != "" {
				// 	return
				// }
				config.SetNotice(v)
			}
		}
	case pb.CONFIG_SHOP: //商城
		b := make(map[string]data.Shop)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}

		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelShop(k)
			case pb.CONFIG_UPSERT:
				config.SetShop(v)
			}
		}
	case pb.CONFIG_GAMES: //游戏
		b := make(map[string]data.Game)
		err = jsoniter.Unmarshal(msg.Data, &b)
		// glog.Debugf("Sync Games %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelGame(k)
			case pb.CONFIG_UPSERT:
				config.SetGame(v)
			}
		}
	case pb.CONFIG_VIP: //游戏
		b := make(map[int]data.Vip)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelVip(k)
			case pb.CONFIG_UPSERT:
				config.SetVip(v)
			}
		}
	case pb.CONFIG_TASK: //任务
		b := make(map[int32]data.Task)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync Task %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelTask(k)
			case pb.CONFIG_UPSERT:
				config.SetTask(v)
			}
		}
	case pb.CONFIG_POKERHANDS: //牌型任务
		b := make(map[int32]data.PokerHands)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync PokerHands task %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelPhTask(k)
			case pb.CONFIG_UPSERT:
				config.SetPhTask(v)
			}
		}
	case pb.CONFIG_LUCKY: //lucky
		b := make(map[int32]data.Lucky)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync Lucky %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelLucky(k)
			case pb.CONFIG_UPSERT:
				config.SetLucky(v)
			}
		}
	case pb.CONFIG_ACT: //activity
		b := make(map[string]data.Activity)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync Activity %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelActivity(k)
			case pb.CONFIG_UPSERT:
				config.SetActivity(v)
			}
		}
	case pb.CONFIG_SIGN: //DailySign
		b := make(map[string]data.DailySign)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync DailySign %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelDailySign(k)
			case pb.CONFIG_UPSERT:
				config.SetDailySign(v)
			}
		}
	case pb.CONFIG_ONLINE: //onlinereward
		b := make(map[string]data.OnlineReward)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync OnlineReward %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelOnlineReward(k)
			case pb.CONFIG_UPSERT:
				config.SetOnlineReward(v)
			}
		}
	case pb.CONFIG_LIMITEDGIFT: //限时礼包
		b := make(map[string]data.LimitedGift)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync LimitedGift %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelLimitedGift(k)
			case pb.CONFIG_UPSERT:
				config.SetLimitedGift(v)
			}
		}
	case pb.CONFIG_LOGIN: //登录奖励
		b := make(map[uint32]data.LoginPrize)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync LoginPrize %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelLogin(k)
			case pb.CONFIG_UPSERT:
				config.SetLogin(v)
			}
		}
	case pb.CONFIG_WITHDRAW: //提现配置
		b := make(map[int32]data.Withdraw)
		err = jsoniter.Unmarshal(msg.Data, &b)
		glog.Debugf("Sync Withdraw %#v", b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelWithDraw(k)
			case pb.CONFIG_UPSERT:
				config.SetWithDraw(v)
			}
		}
	case pb.CONFIG_SWITCH:
		// 功能开关
		b := make(map[string]data.SystemSwitch)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, v := range b {
			switch msg.Atype {
			case pb.CONFIG_DELETE:
				config.DelSystemSwitch(v.Id)
			case pb.CONFIG_UPSERT:
				config.SetSystemSwitch(v)
			}
		}
	case pb.CONFIG_UPLOAD_CONFIG:
		err = updateConfig(msg.Data, node)
		if err != nil {
			glog.Errorf("updateConfig err %v, data %#v", err, msg.Data)
			return
		}
		glog.Info("uploadConfig success")
	case pb.CONFIG_RELOAD:
		err = table.LoadTables()
		if err != nil {
			glog.Errorf("reloadConfig err %v", err)
			return
		}
		glog.Info("reloadConfig success")
	case pb.CONFIG_PAY_CHANNEL:
		// 支付渠道
		b := make(map[string]data.PayChannel)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		channels := make([]data.PayChannel, 0)
		for _, pc := range b {
			channels = append(channels, pc)
		}
		config.SetPayChannel(channels)
	case pb.CONFIG_SHARE:
		b := make(map[string]data.Share)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			config.SetShare(s)
		}
	case pb.CONFIG_SHARE_ADDR:
		b := make(map[string]data.ShareAddr)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			config.SetShareAddr(s)
		}
	case pb.CONFIG_BEGINNER:
		b := make(map[string]data.Beginner)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			config.SetBeginner(s)
		}
	case pb.CONFIG_IPWHITE:
		b := make(map[string]string)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetIpWhite(k)
			} else {
				config.DeleteIpWhite(k)
			}
		}
	case pb.CONFIG_WEEKCARD:
		b := make(map[int32]data.WeeklyCard)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, v := range b {
			config.SetWeeklyCard(v)
		}
	case pb.CONFIG_EMOJI:
		b := make(map[string]data.Emoji)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			config.SetEmoji(s)
		}
	case pb.CONFIG_RECHARGELIMIT:
		b := make(map[string]data.RechargeLimit)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetPhoneBlack(s)
			} else {
				config.DelPhoneBlack(s)
			}
		}
	case pb.CONFIG_CHANNEL:
		b := make(map[string]data.ChannelPackage)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetChannelPackage(s)
			} else {
				config.DeleteChannelPackage(s)
			}
		}
	case pb.CONFIG_SEVERWHITE:
		b := make(map[string]string)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetServerWhite(k)
			} else {
				config.DeleteServerWhite(k)
			}
		}
	case pb.CONFIG_CUSTOMERADDR:
		// 客服地址
		b := make(map[string]data.ModifyCustomer)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetCustomerAddress(v)
			} else {
				config.DelCustomerAddress(k)
			}
		}
	case pb.CONFIG_SHARE_WAY:
		// 分享配置
		b := make(map[string]data.ModifyShare)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for k, v := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetShareWay(v)
			} else {
				config.DelShareWay(k)
			}
		}
	case pb.CONFIG_PVP_ROOM: // 对战房
		pvp := &data.PvpRoom{}
		err = jsoniter.Unmarshal(msg.Data, pvp)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetPvpRoom(pvp)
	case pb.CONFIG_RANK_WITHDRAW: // 提现排行榜
		var ranks []*data.RankWithdraw
		err = jsoniter.Unmarshal(msg.Data, &ranks)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetRankWithdraw(ranks)
	case pb.CONFIG_VIP_ROBOT: // 人机VIP配置
		var vips []*data.VipRobot
		err = jsoniter.Unmarshal(msg.Data, &vips)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetVipRobot(vips)
	case pb.CONFIG_LUCKY_DRAW_REWORD: // 小米手机活动奖励
		var rewords []*data.LuckyDrawReward
		err = jsoniter.Unmarshal(msg.Data, &rewords)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetLuckyDrawRewards(rewords)
	case pb.CONFIG_LUCKY_DRAW_ROBOT: // 小米手机活动自动放号
		var robots []*data.LuckyDrawRobot
		err = jsoniter.Unmarshal(msg.Data, &robots)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetLuckyDrawRobots(robots)
	case pb.CONFIG_DEVICELIMIT:
		b := make(map[string]data.DeviceIdLimit)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, s := range b {
			if msg.Atype == pb.CONFIG_UPSERT {
				config.SetDeviceBlack(s)
			} else {
				config.DeleteDeviceBlack(s.Id)
			}
		}
	case pb.CONFIG_SCRATCHTICK: // 大富翁
		var rewords []*data.ScratchTicketConfig
		err = jsoniter.Unmarshal(msg.Data, &rewords)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetScratchTicketConfig(rewords)
	case pb.CONFIG_MARQUEE_WITHDRAW: // 新跑马灯配置
		var robots []*data.MarqueeWithdraw
		err = jsoniter.Unmarshal(msg.Data, &robots)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetMarqueeWithdraws(robots)
	case pb.CONFIG_PLAYSHARE:
		b := make(map[int]data.PlayAndDrawActivity)
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		for _, v := range b {
			config.SetPlayShareConfig(v)
		}
	case pb.CONFIG_CRASHSTRATEGY:
		b := data.CrashStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetCrashStrategy(b)
	case pb.CHARGE_CLASSIFY:
		b := data.ChargeClassify{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetChargeClassify(b)
	case pb.CONFIG_LHDSTRATEGY:
		b := data.LhdStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetLHDStrategy(b)
	case pb.CONFIG_UPSTRATEGY:
		b := data.SevenStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetSevenStrategy(b)
	case pb.CONFIG_ABSTRATEGY:
		b := data.ABStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetABStrategy(b)
	case pb.CONFIG_CPSTRATEGY:
		b := data.CPStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetCPStrategy(b)
	case pb.CONFIG_RBSTRATEGY:
		b := data.RBStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetRBStrategy(b)
	case pb.CONFIG_AVSTRATEGY:
		b := data.AviatorStrategy{}
		err = jsoniter.Unmarshal(msg.Data, &b)
		if err != nil {
			glog.Errorf("syncConfig Unmarshal err %v, data %#v", err, msg.Data)
			return
		}
		config.SetAviatorStrategy(b)
	default:
		glog.Errorf("syncConfig unknown type %s", msg.Type)
		err = fmt.Errorf("type not exist %d", msg.Type)
	}
	return
}

// 打包配置
func syncConfigMsg(d interface{}) ([]byte, error) {
	result, err := jsoniter.Marshal(d)
	if err != nil {
		glog.Errorf("syncConfig Marshal err %v", err)
	}
	return result, err
}

// SyncConfig2 打包消息
func SyncConfig2(ctype pb.ConfigType, atype pb.ConfigAtype,
	data []byte) (msg *pb.SyncConfig) {
	msg = new(pb.SyncConfig)
	msg.Type = ctype
	msg.Atype = atype
	msg.Data = data
	return
}

// GetSyncConfig2 同步配置
func GetSyncConfig2(ctype pb.ConfigType) (msg *pb.SyncConfig, err error) {
	msg = new(pb.SyncConfig)
	msg.Type = ctype
	msg.Atype = pb.CONFIG_UPSERT
	switch ctype {
	case pb.CONFIG_ENV: //变量
		msg.Data, err = syncConfigMsg(config.GetEnvs())
	case pb.CONFIG_NOTICE: //公告
		msg.Data, err = syncConfigMsg(config.GetNotices2())
	case pb.CONFIG_SHOP: //商城
		msg.Data, err = syncConfigMsg(config.GetShops2())
	case pb.CONFIG_GAMES: //游戏列表
		msg.Data, err = syncConfigMsg(config.GetGames2())
	case pb.CONFIG_VIP: //vip列表
		// msg.Data, err = syncConfigMsg(config.GetVips2())
	case pb.CONFIG_TASK: //任务列表
		msg.Data, err = syncConfigMsg(config.GetTasks2())
	case pb.CONFIG_POKERHANDS: //牌型任务列表
		msg.Data, err = syncConfigMsg(config.GetPhTasks2())
	case pb.CONFIG_LUCKY: //lucky列表
		msg.Data, err = syncConfigMsg(config.GetLuckys2())
	case pb.CONFIG_ACT: //activity列表
		msg.Data, err = syncConfigMsg(config.GetActivitys2())
	case pb.CONFIG_SIGN: //签到
		msg.Data, err = syncConfigMsg(config.GetDailySigns2())
	case pb.CONFIG_ONLINE: //在线奖励
		msg.Data, err = syncConfigMsg(config.GetOnlineReward2())
	case pb.CONFIG_LIMITEDGIFT: //限时礼包
		msg.Data, err = syncConfigMsg(config.GetLimitedGift2())
	case pb.CONFIG_LOGIN: //登录奖励列表
		msg.Data, err = syncConfigMsg(config.GetLogins2())
	case pb.CONFIG_WITHDRAW: //提现配置
		msg.Data, err = syncConfigMsg(config.GetWithDraw2())
	case pb.CONFIG_SWITCH: //系统开关
		msg.Data, err = syncConfigMsg(config.GetSettingMap())
	case pb.CONFIG_PAY_CHANNEL: //充值渠道
		msg.Data, err = syncConfigMsg(config.GetPayChannelMap())
	case pb.CONFIG_SHARE: //分享
		msg.Data, err = syncConfigMsg(config.GetShareMap())
	case pb.CONFIG_SHARE_ADDR: //分享链接
		msg.Data, err = syncConfigMsg(config.GetShareAddrMap())
	case pb.CONFIG_BEGINNER: //新手配置
		msg.Data, err = syncConfigMsg(config.GetBeginnerMap())
	case pb.CONFIG_IPWHITE: //ip白名单
		msg.Data, err = syncConfigMsg(config.GetIpWhiteMap())
	case pb.CONFIG_WEEKCARD: //周卡
		// msg.Data, err = syncConfigMsg(config.GetWeeklyCardMap())
	case pb.CONFIG_EMOJI: //人机表情
		msg.Data, err = syncConfigMsg(config.GetEmojiMap())
	case pb.CONFIG_RECHARGELIMIT: //充值黑名单
		msg.Data, err = syncConfigMsg(config.GetPhoneBlackMap())
	case pb.CONFIG_CHANNEL: //渠道包
		msg.Data, err = syncConfigMsg(config.GetChannelPackageMap())
	case pb.CONFIG_SEVERWHITE: //ip白名单
		msg.Data, err = syncConfigMsg(config.GetServerWhiteMap())
	case pb.CONFIG_CUSTOMERADDR: //客服地址
		msg.Data, err = syncConfigMsg(config.GetCustomerAddressMap())
	case pb.CONFIG_SHARE_WAY: //分享配置
		msg.Data, err = syncConfigMsg(config.GetShareWayMap())
	case pb.CONFIG_PVP_ROOM: // 对战房
		msg.Data, err = syncConfigMsg(config.GetPvpRoom())
	case pb.CONFIG_RANK_WITHDRAW: // 提现排行榜
		msg.Data, err = syncConfigMsg(config.GetRankWithdraw())
	case pb.CONFIG_VIP_ROBOT: // 人机VIP配置
		msg.Data, err = syncConfigMsg(config.GetVipRobot())
	case pb.CONFIG_LUCKY_DRAW_REWORD: // 小米手机活动奖励
		msg.Data, err = syncConfigMsg(config.GetLuckyDrawRewards())
	case pb.CONFIG_LUCKY_DRAW_ROBOT: // 小米手机活动自动放号
		msg.Data, err = syncConfigMsg(config.GetLuckyDrawRobots())
	case pb.CONFIG_DEVICELIMIT: //设备码黑名单
		msg.Data, err = syncConfigMsg(config.GetDeviceBlackMap())
	case pb.CONFIG_SCRATCHTICK: // 大富翁
		msg.Data, err = syncConfigMsg(config.GetScratchTicketConfig())
	case pb.CONFIG_MARQUEE_WITHDRAW: // 新跑马灯
		msg.Data, err = syncConfigMsg(config.GetMarqueeWithdraws())
	case pb.CONFIG_PLAYSHARE: //玩游戏分享
		msg.Data, err = syncConfigMsg(config.GetPlayShareConfigMap())
	case pb.CONFIG_CRASHSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetCrashStrategy())
	case pb.CONFIG_LHDSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetLHDStrategy())
	case pb.CHARGE_CLASSIFY:
		msg.Data, err = syncConfigMsg(config.GetChargeClassify())
	case pb.CONFIG_UPSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetSevenStrategy())
	case pb.CONFIG_ABSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetABStrategy())
	case pb.CONFIG_CPSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetCPStrategy())
	case pb.CONFIG_RBSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetRBStrategy())
	case pb.CONFIG_AVSTRATEGY:
		msg.Data, err = syncConfigMsg(config.GetAviatorStrategy())
	default:
		err = fmt.Errorf("type not exist %d", ctype)
	}
	return
}

func SyncConfig3(ctype pb.ConfigType) (msg *pb.SyncConfig) {
	msg = new(pb.SyncConfig)
	msg.Type = ctype
	return
}

// GetSyncConfig 获取
func GetSyncConfig(ctype pb.ConfigType) (msg *pb.SyncConfig) {
	msg, err := GetSyncConfig2(ctype)
	if err != nil {
		glog.Errorf("err %s", err)
	}
	return
}
