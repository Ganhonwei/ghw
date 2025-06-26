package event

import (
	"goserver/pkg/data"
	"goserver/pkg/glog"

	jsoniter "github.com/json-iterator/go"
)

const (
	POKER_HANDS          int32 = iota // 牌型任务
	RECHARGE_TASK                     // 充值任务
	PLAY_TASK                         // 玩某个游戏x局
	RECHARGE_MAIL                     // 充值完成邮件
	WITHDRAW_Mail                     // 提现邮件
	GIVE_CASH                         // 后台充值
	LIMITED_GIFT                      // 限时礼包
	CHECK_STATE                       // 状态检测
	RECHARGE_JAR                      // 充值罐子
	POT_FLOW                          // 罐子打码量
	VIP                               // vip充值
	PLAY_SHARE                        // 玩游戏分享
	Normal_JAR                        // 赠送罐子
	Initiative_Leave                  // crash主动离开
	Crash_FYZS                        // crash扶摇直上生效
	Crash_RKYH_Tigger                 // crash人狂有祸触发
	Crash_Settlement                  // crash结算
	Crash_QSHS                        // crash起死回生生效
	Crash_MXJL                        // crash冒险奖励生效
	CRASH_JACKPOT                     // crash奖池
	CRASH_GCYX_SCORE                  // crash高潮涌现输赢分
	CRASH_GCYX_BET                    // crash高潮涌现下注
	CRASH_GCYX_Trigger                // crash高潮涌现触发
	CRASH_GCXY_GC_Over                // crash高潮涌现高潮结束
	CRASH_GCXY_XZ_Over                // crash高潮涌现贤者结束
	Aviator_FYZS                      // Aviator扶摇直上生效
	Aviator_RKYH_Tigger               // Aviator人狂有祸触发
	Aviator_Settlement                // Aviator结算
	Aviator_QSHS                      // Aviator起死回生生效
	Aviator_MXJL                      // Aviator冒险奖励生效
	Aviator_JACKPOT                   // Aviator奖池
	Aviator_GCYX_SCORE                // Aviator高潮涌现输赢分
	Aviator_GCYX_BET                  // Aviator高潮涌现下注
	Aviator_GCYX_Trigger              // Aviator高潮涌现触发
	Aviator_GCXY_GC_Over              // Aviator高潮涌现高潮结束
	Aviator_GCXY_XZ_Over              // Aviator高潮涌现贤者结束
	LHD_Strategy                      // lhd策略生效
	LHD_XXSC_Strategy                 // lhd心想事成策略生效
	LHD_LKYH_Trigger                  // lhd策略触发
	LHD_LKYH_BS                       // lhd触发倍杀
	LHD_Reset                         // lhd重置
	Seven_Strategy                    // seven策略生效
	Seven_XXSC_Strategy               // seven心想事成策略生效
	Seven_QGY_Trigger                 // seven7管严策略触发
	VB_GAME_TASK                      // vip bank 游戏任务
	AB_Strategy                       // ab策略生效
	CP_Strategy                       // 彩票策略生效
	RB_Strategy                       // 红黑策略生效
	BreakingGift                      // 破产礼包
)

func Event(user *data.User, event int32, target any) any {
	switch event {
	case POKER_HANDS:
		if p, ok := target.(*PokerHandsEvent); ok {
			return p.event(user)
		}
	case RECHARGE_TASK:
		if p, ok := target.(*RechargeTaskEvent); ok {
			return p.event(user)
		}
	case PLAY_TASK:
		if p, ok := target.(*GameRecordEvent); ok {
			return p.event(user)
		}
	case RECHARGE_MAIL:
		if p, ok := target.(*RechargeMailEvent); ok {
			return p.event(user)
		}
	case WITHDRAW_Mail:
		if p, ok := target.(*WithdrawMailEvent); ok {
			return p.event(user)
		}
	case GIVE_CASH:
		if p, ok := target.(*GiveCash); ok {
			return p.event(user)
		}
	case LIMITED_GIFT:
		if p, ok := target.(*LimitedGiftEvent); ok {
			return p.event(user)
		}
	case CHECK_STATE:
		if p, ok := target.(*CheckStateEvent); ok {
			return p.event(user)
		}
	case RECHARGE_JAR:
		if p, ok := target.(*RechargeJarEvent); ok {
			return p.event(user)
		}
	case VIP:
		if p, ok := target.(*VIPEvent); ok {
			return p.event(user)
		}
	case PLAY_SHARE:
		if p, ok := target.(*PlayShareEvent); ok {
			return p.event(user)
		}
	case Normal_JAR:
		if p, ok := target.(*NormalJarEvent); ok {
			return p.event(user)
		}
	case LHD_Reset:
		if p, ok := target.(*LHDStrategyResetEvent); ok {
			return p.event(user)
		}
	}
	glog.Errorf("event fail, id:%d, user:%s", user.Userid)
	return nil
}

func EventPost(user *data.User, event int32, data []byte) any {
	switch event {
	case POKER_HANDS:
		target := new(PokerHandsEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case PLAY_TASK:
		target := new(GameRecordEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case POT_FLOW:
		target := new(ShopPotFlowEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Initiative_Leave:
		target := new(InitiativeLeaveEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Crash_FYZS:
		target := new(CrashFYZSEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Crash_RKYH_Tigger:
		target := new(CrashRKYHTiggerEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Crash_Settlement:
		target := new(CrashSettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Crash_QSHS:
		target := new(CrashQSHSEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Crash_MXJL:
		target := new(CrashMXJLEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_FYZS:
		target := new(AviatorFYZSEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_RKYH_Tigger:
		target := new(AviatorRKYHTiggerEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_Settlement:
		target := new(AviatorSettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_QSHS:
		target := new(AviatorQSHSEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_MXJL:
		target := new(AviatorMXJLEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_JACKPOT:
		target := new(AviatorJackpotEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_GCYX_Trigger:
		target := &AviatorGCYXTriggerEvent{}
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_GCYX_SCORE:
		target := new(AviatorGCYXScoreEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_GCYX_BET:
		target := new(AviatorGCYXBetEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_GCXY_GC_Over:
		target := new(AviatorGCXYGCOverEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Aviator_GCXY_XZ_Over:
		target := new(AviatorGCXYXZOverEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case LHD_Strategy:
		target := new(LHDStrategySettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case LHD_XXSC_Strategy:
		target := new(LHDXXSCSettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Seven_Strategy:
		target := new(SevenStrategySettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Seven_XXSC_Strategy:
		target := new(SevenXXSCSettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case Seven_QGY_Trigger:
		target := new(SevenQGYTriggerEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case VB_GAME_TASK:
		return nil
	// 关闭vb任务
	// 	target := new(VBGameTaskEvent)
	// 	err := jsoniter.Unmarshal(data, target)
	// 	if err != nil {
	// 		glog.Errorf("event post fail, eventId:%d", event)
	// 		return nil
	// 	}
	// 	return target.event(user)
	case CRASH_JACKPOT:
		target := new(CrashJackpotEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CRASH_GCYX_Trigger:
		target := &CrashGCYXTriggerEvent{}
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CRASH_GCYX_SCORE:
		target := new(CrashGCYXScoreEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CRASH_GCYX_BET:
		target := new(CrashGCYXBetEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CRASH_GCXY_GC_Over:
		target := new(CrashGCXYGCOverEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CRASH_GCXY_XZ_Over:
		target := new(CrashGCXYXZOverEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case LHD_LKYH_BS:
		target := new(LHDLKYHBSEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case LHD_LKYH_Trigger:
		target := new(LHDLKYHTriggerEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case AB_Strategy:
		target := new(ABStrategySettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case CP_Strategy:
		target := new(CPStrategySettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case RB_Strategy:
		target := new(RBStrategySettlementEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	case BreakingGift:
		target := new(BreakingGiftEvent)
		err := jsoniter.Unmarshal(data, target)
		if err != nil {
			glog.Errorf("event post fail, eventId:%d", event)
			return nil
		}
		return target.event(user)
	}
	glog.Errorf("event fail, id:%d, user:%s", user.Userid)
	return nil
}
