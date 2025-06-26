package event

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"sort"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"gopkg.in/mgo.v2/bson"
)

// 牌型任务事件触发
func (p *PokerHandsEvent) event(user *data.User) proto.Message {
	// if !p.Win || p.PX != algo.BaoZi {
	// 	return nil
	// }
	// glog.Infof("PX add progress, id:%s", user.Userid)
	// user.PhProgress += 1
	// task := config.GetPhTask(user.PhRewardId)
	// ntf := new(pb.PokerHandsTaskNtf)
	// ntf.Progress = user.PhProgress
	// ntf.MaxProgress = task.Progress
	// return ntf
	return nil
}

// 充值任务事件触发
func (p *RechargeTaskEvent) event(user *data.User) proto.Message {
	rsp := new(pb.TaskProgressNtf)
	taskMap := user.Task
	for _, task := range taskMap {
		if task.TaskType != int32(pb.TASK_TYPE1) || task.Prize != 1 {
			continue
		}
		task.Num += p.Amount
		if handler.IsFinishTask(task.Taskid, task.Num) {
			task.Prize = 2
		}
		bean, err := handler.BuildTaskData(task)
		if err != nil {
			continue
		}
		rsp.Task = append(rsp.Task, bean)
	}
	return rsp
}

// 游戏结算触发
func (g *GameRecordEvent) event(user *data.User) proto.Message {
	rsp := new(pb.TaskProgressNtf)
	taskMap := user.Task
	for _, task := range taskMap {
		if task.Prize != 1 {
			continue
		}
		t := config.GetTask(task.Taskid)
		if t.ID == 0 {
			continue
		}
		switch task.TaskType {
		case int32(pb.TASK_TYPE2):
			// 玩X类型游戏N局
			if t.Count1 != g.Gtype && t.Count1 != 0 {
				continue
			}
			task.Num += 1
		case int32(pb.TASK_TYPE3):
			// 赢X类型游戏N局
			if (t.Count1 != g.Gtype && t.Count1 != 0) || !g.Win {
				continue
			}
			task.Num += 1
		default:
			continue
		}
		// 判断完没完成
		if handler.IsFinishTask(task.Taskid, task.Num) {
			task.Prize = 2
		}
		bean, err := handler.BuildTaskData(task)
		if err != nil {
			continue
		}
		rsp.Task = append(rsp.Task, bean)
	}
	if len(rsp.Task) <= 0 {
		return nil
	}
	return rsp
}

// 充值完成事件触发
func (p *RechargeMailEvent) event(user *data.User) proto.Message {
	ntf := new(pb.FeedBackLogNtf)
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: user.Userid,
		Title:    "System Message",
		Content:  handler.BuildRechargeFinish(int64(p.Amount), user.Nickname),
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	user.FeedBackLogMap[chat.Uid] = chat
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	return ntf
}

// 提现成功事件触发
func (p *WithdrawMailEvent) event(user *data.User) proto.Message {
	content := ""
	switch p.Status {
	case data.WithdrawSuccess:
		content = handler.BuildWithdrawSuccess(user.Nickname)
	case data.WithdrawFail:
		content = handler.BuildWithdrawFail(user.Nickname)
	case data.WithdrawBack:
		content = handler.BuildWithdrawBack(user.Nickname)
	// case data.WithdrawBackBankErr:
	// 	content = handler.BuildWithdrawBackBank(user.Nickname)
	default:
		return nil
	}
	ntf := new(pb.FeedBackLogNtf)
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: user.Userid,
		Title:    "System Message",
		Content:  content,
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	user.FeedBackLogMap[chat.Uid] = chat
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	return ntf
}

// 后台充值
func (p *GiveCash) event(user *data.User) proto.Message {
	ntf := new(pb.FeedBackLogNtf)
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: user.Userid,
		Title:    "System Message",
		Content:  handler.BuildGiveCash(int64(p.Amount), user.Nickname),
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	user.FeedBackLogMap[chat.Uid] = chat
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	return ntf
}

func (e *LimitedGiftEvent) event(user *data.User) proto.Message {
	gifts := config.GetLimitedGiftList()

	var ntf *pb.LimitedGiftNtf
	if user.LimitedGiftId != 0 && user.LimitedGiftOverTime > utils.LocalTime().UnixMilli() {
		// user存在礼包
		return nil
	}
	for _, g := range gifts {
		if g.Switch == 0 {
			// 礼包没开
			continue
		}

		ov := false
		overs := user.OverlimitedGift // 买过或过期的礼包
		for _, v := range overs {
			if v == g.Id {
				ov = true
				break
			}
		}
		if ov {
			// 已经触发过了
			continue
		}

		if user.LimitedGiftId != 0 && user.LimitedGiftOverTime > utils.LocalTime().UnixMilli() && g.GameRound <= int32(user.Round) && g.Recharge <= int32(user.Money) {
			// user存在礼包
			user.OverlimitedGift = append(user.OverlimitedGift, g.Id)
			continue
		}

		if g.GameRound <= int32(user.Round) && g.Recharge <= int32(user.Money) {
			user.LimitedGiftId = g.Id
			user.LimitedGiftOverTime = utils.LocalTime().UnixMilli() + g.DuringTime*1000
			ntf = &pb.LimitedGiftNtf{
				Id:       g.Id,
				Price:    g.Price,
				Cash:     int32(g.Reward),
				OverTime: user.LimitedGiftOverTime,
			}
		}
	}

	if ntf != nil {
		return ntf
	}
	return nil
}

func (e *CheckStateEvent) event(user *data.User) proto.Message {
	if handler.NormalToFrothState(user) {
		user.State = data.FrothState
		return &pb.UserGameState{
			Userid: user.Userid,
			State:  int32(user.State),
		}
	}

	return nil
}

func (r *RechargeJarEvent) event(user *data.User) proto.Message {
	return nil
	// var multiple, give int64 = 40, r.Give // 默认40倍

	// if r.ShopType == data.METAL_CARD {
	// 	// 周卡不给罐子,主要针对C类用户
	// 	return nil
	// }

	// if r.ShopType > 0 && r.ShopType != data.SHOP && user.RegistArea != 2 {
	// 	// 非C类用户只有商城充值才会触发罐子，C类用户充值赠送都会进罐子
	// 	return nil
	// }

	// if r.ShopType == data.SHOP {
	// 	shop := config.GetShop(r.ShopId)

	// 	giveType := 2
	// 	if len(shop.GiveType) >= user.RegistArea+1 {
	// 		giveType = shop.GiveType[user.RegistArea]
	// 	}
	// 	if len(shop.Give) >= user.RegistArea+1 {
	// 		give = shop.Give[user.RegistArea]
	// 	}
	// 	if len(shop.FlowMultiple) >= user.RegistArea+1 {
	// 		multiple = int64(shop.FlowMultiple[user.RegistArea])
	// 	}

	// 	if shop.Id == "" || giveType != 2 {
	// 		return nil
	// 	}

	// 	if multiple == 0 {
	// 		// 没配置,默认40倍
	// 		multiple = 40
	// 	}
	// }

	// // if user.RegistArea != 2 {
	// // 	// 非C类用户

	// // }

	// if give <= 0 {
	// 	return nil
	// }

	// now := utils.BsonNow().Unix()
	// flow := multiple * give
	// build := data.ShopPot{
	// 	Id:        bson.NewObjectId().String(),
	// 	Number:    give,
	// 	FlowWater: int64(flow),
	// 	CTime:     now,
	// 	OverTime:  now + 30*24*3600,
	// }
	// user.ActivatedPot = append(user.ActivatedPot, build)
	// ntf := &pb.ActivityDataNtf{
	// 	ShopPots: handler.BuildShopPotDataMsg(user),
	// }
	// return ntf
}

func (s *ShopPotFlowEvent) event(user *data.User) proto.Message {
	if len(user.ActivatedPot) <= 0 {
		return nil
	}

	var maxProgress int64
	pots := user.ActivatedPot
	for _, p := range pots {
		maxProgress += p.FlowWater
	}

	if user.PotFlow >= maxProgress {
		return nil
	}

	user.PotFlow += int64(math.Abs(float64(s.Score)))
	if user.PotFlow > maxProgress {
		user.PotFlow = maxProgress
	}

	return &pb.ActivityDataNtf{
		PotsFlow: user.PotFlow,
	}
}

func (v *VIPEvent) event(user *data.User) proto.Message {
	handler.AddVipExp(user, v.Amount)

	return &pb.UpdateVipExpNtf{
		Lv:       int32(user.Vip.Lv),
		Exp:      user.Vip.Exp,
		BeforeLv: int32(user.Vip.BeforeLv),
	}
}

func (v *PlayShareEvent) event(user *data.User) proto.Message {
	c := config.GetPlayShareConfig(1)
	if c.Id == 0 {
		return nil
	}
	if user.PlayShareData.Config.Id == 0 {
		user.PlayShareData.Config = c
	}

	c = user.PlayShareData.Config

	// 判断活动开没开
	if !config.SettingIsOpenByName(1, "拼多多") {
		return nil
	}

	data := user.PlayShareData
	if data.OverTime == 0 {
		user.PlayShareData.OverTime = utils.BsonNow().Unix() + int64(c.ValidityTime*60*60)
	}

	if user.PlayShareData.OverTime < utils.BsonNow().Unix() || data.Over {
		return nil
	}

	if v.Ptype == 1 {
		user.PlayShareData.RealPlayTimes++
		user.PlayShareData.PlayTimes++
	} else {
		add := v.DeviceId != ""
		for _, v2 := range data.InviteFriends {
			if v2 == v.DeviceId {
				add = false
				break
			}
		}
		if add {
			user.PlayShareData.RealInviteTimes++
			user.PlayShareData.InviteFriends = append(user.PlayShareData.InviteFriends, v.DeviceId)
		}
	}

	// 计算是否能获得抽奖次数
	// 玩游戏的
	if user.PlayShareData.RealPlayTimes >= int(c.PlayRounds[0]) && data.PlayUsed+data.PlayDraws < int(c.MaxPlayDraw) {
		// 加次数
		addTimes := user.PlayShareData.RealPlayTimes / int(c.PlayRounds[0])
		user.PlayShareData.PlayDraws += addTimes * int(c.PlayRounds[1])
		user.PlayShareData.RealPlayTimes -= addTimes * int(c.PlayRounds[0])
		user.PlayShareData.PlayDraws = int(math.Min(float64(user.PlayShareData.PlayDraws), float64(c.MaxPlayDraw)))
	}
	// 分享的
	if user.PlayShareData.RealInviteTimes >= int(c.ShareFriends[0]) && data.ShareUsed+data.ShareDraws < int(c.MaxShareDraw) {
		// 加次数
		addTimes := user.PlayShareData.RealInviteTimes / int(c.ShareFriends[0])
		user.PlayShareData.ShareDraws += addTimes * int(c.ShareFriends[1])
		user.PlayShareData.RealInviteTimes -= addTimes * int(c.ShareFriends[0])
		user.PlayShareData.ShareDraws = int(math.Min(float64(user.PlayShareData.ShareDraws), float64(c.MaxShareDraw)))
	}

	return &pb.ActivityDataNtf{
		PlayShare: handler.BuildPlayShareDataMsg(user),
	}
}

func (r *NormalJarEvent) event(user *data.User) proto.Message {
	if user == nil {
		glog.Errorf("user is nil")
		return nil
	}
	multiple := int64(r.Multiple)

	if multiple <= 0 {
		multiple = 40
	}

	if r.Give <= 0 {
		return nil
	}

	now := utils.BsonNow().Unix()
	flow := multiple * r.Give
	build := data.ShopPot{
		Id:        bson.NewObjectId().String(),
		Number:    r.Give,
		FlowWater: int64(flow),
		CTime:     now,
		OverTime:  now + 30*24*3600,
	}
	if user.ActivatedPot == nil {
		user.ActivatedPot = make([]data.ShopPot, 0)
	}
	user.ActivatedPot = append(user.ActivatedPot, build)
	ntf := &pb.ActivityDataNtf{
		ShopPots: handler.BuildShopPotDataMsg(user),
	}
	return ntf
}

func (r *InitiativeLeaveEvent) event(user *data.User) proto.Message {
	switch r.Gtype {
	case int(pb.CRASH):
		if user.CrashStrategy.FYZS.EvoTimes > 0 {
			user.CrashStrategy.FYZS.EvoTimes = 0
			// user.CrashStrategy.FYZS.PlayTimes++
			user.CrashStrategy.FYZS.AllTiggerTimes++
		}
	case int(pb.PLANE):
		if user.AvStrategy.FYZS.EvoTimes > 0 {
			user.AvStrategy.FYZS.EvoTimes = 0
			user.AvStrategy.FYZS.AllTiggerTimes++
		}
	case int(pb.LHD):
		if user.LHDStrategy.XXSC.Evo {
			user.LHDStrategy.XXSC.MaxTriggerTimes++
		}
	case int(pb.SEVEN):
		if user.SevenStrategy.XXSC.Evo {
			user.SevenStrategy.XXSC.MaxTriggerTimes++
		}
	case int(pb.ABAR):
		if user.ABStrategy.ARTY.EvoTimes {
			user.ABStrategy.ARTY.UZ++
		}
	case int(pb.LOTTERY):
		if user.CPStrategy.LWJY.EvoTimes {
			user.CPStrategy.LWJY.UZ++
		}
	case int(pb.REDBLACK):
		if user.RBStrategy.HYDT.EvoTimes {
			user.RBStrategy.HYDT.UZ++
		}
	}
	return nil
}

func (r *CrashFYZSEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.FYZS.EvoTimes++
	user.CrashStrategy.FYZS.AllEvoTimes++
	user.CrashStrategy.FYZS.WinScore += r.Score

	s := config.GetCrashStrategy()
	ctype := handler.GetChargeType(user)
	if user.Diamond >= s.FYZS.M[ctype] {
		user.CrashStrategy.FYZS.TiggerTimes++
	}
	return nil
}

func (r *CrashRKYHTiggerEvent) event(user *data.User) proto.Message {
	if r.Tigger {
		user.CrashStrategy.RKYH.TriggerTimes++
	}
	user.CrashStrategy.RKYH.WinMulpitles = append(user.CrashStrategy.RKYH.WinMulpitles, r.WinMulti)
	length := len(user.CrashStrategy.RKYH.WinMulpitles)
	w := r.W
	if w <= 0 {
		w = 10
	}
	if length > int(w) {
		user.CrashStrategy.RKYH.WinMulpitles = user.CrashStrategy.RKYH.WinMulpitles[length-int(w):]
	}
	return nil
}

// Deprecated
func (r *CrashSettlementEvent) event(user *data.User) proto.Message {
	bean := config.GetCrashStrategy()

	if bean.Id != 0 {
		ctype := handler.GetChargeType(user)

		// 欲薅无门策略
		if user.CrashStrategy.YHWM.Tigger {
			if r.Score < 0 {
				user.CrashStrategy.YHWM.RecyleScore -= r.Score
				user.CrashStrategy.YHWM.AllRecyleScore -= r.Score
			}

			// 判断是否要取消
			if float64(user.CrashStrategy.YHWM.RecyleScore) >= float64(user.CrashStrategy.YHWM.WinScore)*bean.YHWM.X[ctype] {
				user.CrashStrategy.YHWM = data.CrashYHWM{
					TiggerTimes:    user.CrashStrategy.YHWM.TiggerTimes,
					AllRecyleScore: user.CrashStrategy.YHWM.AllRecyleScore,
				}
			}
		} else {
			// 没生效
			if r.Score > 0 {
				// 策略生效前赢分
				user.CrashStrategy.YHWM.MonitorRounds++
				user.CrashStrategy.YHWM.CrashMulpitles = append(user.CrashStrategy.YHWM.CrashMulpitles, r.Multiple)
				user.CrashStrategy.YHWM.WinScore += r.Score
			} else {
				// 重置
				user.CrashStrategy.YHWM = data.CrashYHWM{
					TiggerTimes:    user.CrashStrategy.YHWM.TiggerTimes,
					AllRecyleScore: user.CrashStrategy.YHWM.AllRecyleScore,
				}
			}

			yhwm := user.CrashStrategy.YHWM
			if (yhwm.TiggerTimes < bean.YHWM.T[ctype] || bean.YHWM.T[ctype] == -1) &&
				yhwm.MonitorRounds >= bean.YHWM.R[ctype] && len(yhwm.CrashMulpitles) > 0 && // 监控局数满足条件
				handler.TiggerCrashStrategy(user, bean.YHWM.BaseStrategy) {
				// 判断是否要触发
				// 逃跑中位数
				sort.Slice(yhwm.CrashMulpitles, func(i, j int) bool {
					return yhwm.CrashMulpitles[i] > yhwm.CrashMulpitles[j]
				})
				middle := float64(yhwm.CrashMulpitles[len(yhwm.CrashMulpitles)/2]) / 100

				// 平均数
				var max int32
				for _, v := range yhwm.CrashMulpitles {
					max += v
				}
				avg := float64(max/int32(len(yhwm.CrashMulpitles))) / 100
				if middle <= bean.YHWM.M[ctype] && (avg <= bean.YHWM.A[ctype] || bean.YHWM.A[ctype] == -1) {
					user.CrashStrategy.YHWM.Tigger = true
					user.CrashStrategy.YHWM.TiggerTimes++
					// user.CrashStrategy.YHWM.AllRecyleScore += user.CrashStrategy.YHWM.RecyleScore
					user.CrashStrategy.YHWM.RecyleScore = 0
				}
			}
		}

		// 起死回生
		if r.Score > 0 {
			user.CrashStrategy.QSHS.WinRounds = append(user.CrashStrategy.QSHS.WinRounds, r.Multiple)
			if len(user.CrashStrategy.QSHS.WinRounds) > 100 {
				user.CrashStrategy.QSHS.WinRounds = user.CrashStrategy.QSHS.WinRounds[len(user.CrashStrategy.QSHS.WinRounds)-100:]
			}
		}

		// 奖池风控
		if r.Multiple > 2000 && r.Score > 0 {
			user.CrashStrategy.JCFK.TriggerTimes++
		}

		// 冒险奖励
		if r.Bet > 0 {
			user.CrashStrategy.MXJL.Bets = append(user.CrashStrategy.MXJL.Bets, r.Bet)
			if len(user.CrashStrategy.MXJL.Bets) > 100 {
				user.CrashStrategy.MXJL.Bets = user.CrashStrategy.MXJL.Bets[len(user.CrashStrategy.MXJL.Bets)-100:]
			}
		}
	}

	return nil
}

func (r *CrashQSHSEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.QSHS.TriggerTimes++
	return nil
}

func (r *CrashMXJLEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.MXJL.TriggerTimes++
	return nil
}

func (r *AviatorFYZSEvent) event(user *data.User) proto.Message {
	user.AvStrategy.FYZS.EvoTimes++
	user.AvStrategy.FYZS.AllEvoTimes++
	user.AvStrategy.FYZS.WinScore += r.Score

	s := config.GetAviatorStrategy()
	ctype := handler.GetChargeType(user)
	if user.Diamond >= s.FYZS.M[ctype] {
		user.AvStrategy.FYZS.TiggerTimes++
	}
	return nil
}

func (r *AviatorRKYHTiggerEvent) event(user *data.User) proto.Message {
	if r.Tigger {
		user.AvStrategy.RKYH.TriggerTimes++
	}
	user.AvStrategy.RKYH.WinMulpitles = append(user.AvStrategy.RKYH.WinMulpitles, r.WinMulti)
	length := len(user.AvStrategy.RKYH.WinMulpitles)
	w := r.W
	if w <= 0 {
		w = 10
	}
	if length > int(w) {
		user.AvStrategy.RKYH.WinMulpitles = user.AvStrategy.RKYH.WinMulpitles[length-int(w):]
	}
	return nil
}

func (r *AviatorSettlementEvent) event(user *data.User) proto.Message {
	bean := config.GetAviatorStrategy()

	if bean.Id != 0 {
		ctype := handler.GetChargeType(user)

		// 欲薅无门策略
		if user.AvStrategy.YHWM.Tigger {
			if r.Score < 0 {
				user.AvStrategy.YHWM.RecyleScore -= r.Score
				user.AvStrategy.YHWM.AllRecyleScore -= r.Score
			}

			// 判断是否要取消
			if float64(user.AvStrategy.YHWM.RecyleScore) >= float64(user.AvStrategy.YHWM.WinScore)*bean.YHWM.X[ctype] {
				user.AvStrategy.YHWM = data.AviatorYHWM{
					TiggerTimes:    user.AvStrategy.YHWM.TiggerTimes,
					AllRecyleScore: user.AvStrategy.YHWM.AllRecyleScore,
				}
			}
		} else {
			// 没生效
			if r.Score > 0 {
				// 策略生效前赢分
				user.AvStrategy.YHWM.MonitorRounds++
				user.AvStrategy.YHWM.AviatorMulpitles = append(user.AvStrategy.YHWM.AviatorMulpitles, r.Multiple)
				user.AvStrategy.YHWM.WinScore += r.Score
			} else {
				// 重置
				user.AvStrategy.YHWM = data.AviatorYHWM{
					TiggerTimes:    user.AvStrategy.YHWM.TiggerTimes,
					AllRecyleScore: user.AvStrategy.YHWM.AllRecyleScore,
				}
			}

			yhwm := user.AvStrategy.YHWM
			if (yhwm.TiggerTimes < bean.YHWM.T[ctype] || bean.YHWM.T[ctype] == -1) &&
				yhwm.MonitorRounds >= bean.YHWM.R[ctype] && len(yhwm.AviatorMulpitles) > 0 && // 监控局数满足条件
				handler.TiggerCrashStrategy(user, bean.YHWM.BaseStrategy) {
				// 判断是否要触发
				// 逃跑中位数
				sort.Slice(yhwm.AviatorMulpitles, func(i, j int) bool {
					return yhwm.AviatorMulpitles[i] > yhwm.AviatorMulpitles[j]
				})
				middle := float64(yhwm.AviatorMulpitles[len(yhwm.AviatorMulpitles)/2]) / 100

				// 平均数
				var max int32
				for _, v := range yhwm.AviatorMulpitles {
					max += v
				}
				avg := float64(max/int32(len(yhwm.AviatorMulpitles))) / 100
				if middle <= bean.YHWM.M[ctype] && (avg <= bean.YHWM.A[ctype] || bean.YHWM.A[ctype] == -1) {
					user.AvStrategy.YHWM.Tigger = true
					user.AvStrategy.YHWM.TiggerTimes++
					user.AvStrategy.YHWM.RecyleScore = 0
				}
			}
		}

		// 起死回生
		if r.Score > 0 {
			user.AvStrategy.QSHS.WinRounds = append(user.AvStrategy.QSHS.WinRounds, r.Multiple)
			if len(user.AvStrategy.QSHS.WinRounds) > 100 {
				user.AvStrategy.QSHS.WinRounds = user.AvStrategy.QSHS.WinRounds[len(user.AvStrategy.QSHS.WinRounds)-100:]
			}
		}

		// 奖池风控
		if r.Multiple > 2000 && r.Score > 0 {
			user.AvStrategy.JCFK.TriggerTimes++
		}

		// 冒险奖励
		if r.Bet > 0 {
			user.AvStrategy.MXJL.Bets = append(user.AvStrategy.MXJL.Bets, r.Bet)
			if len(user.AvStrategy.MXJL.Bets) > 100 {
				user.AvStrategy.MXJL.Bets = user.AvStrategy.MXJL.Bets[len(user.AvStrategy.MXJL.Bets)-100:]
			}
		}
	}

	return nil
}

func (r *AviatorQSHSEvent) event(user *data.User) proto.Message {
	user.AvStrategy.QSHS.TriggerTimes++
	return nil
}

func (r *AviatorMXJLEvent) event(user *data.User) proto.Message {
	user.AvStrategy.MXJL.TriggerTimes++
	return nil
}

func (p *AviatorJackpotEvent) event(user *data.User) proto.Message {
	if p.Jackpot <= 0 {
		return nil
	}

	user.AvStrategy.JCFK.JackpotVal += p.Jackpot
	return nil
}

func (p *AviatorGCYXTriggerEvent) event(user *data.User) proto.Message {
	user.AvStrategy.GCYX.IsStateGC = true
	user.AvStrategy.GCYX.M++
	user.AvStrategy.GCYX.DailyGCTimes++
	user.AvStrategy.GCYX.AvgBet = p.AvgBet
	return nil
}

func (p *AviatorGCYXScoreEvent) event(user *data.User) proto.Message {
	user.AvStrategy.GCYX.R++
	user.AvStrategy.GCYX.WinScore += p.Score
	return nil
}

func (p *AviatorGCYXBetEvent) event(user *data.User) proto.Message {
	user.AvStrategy.GCYX.DMRecord = append(user.AvStrategy.GCYX.DMRecord, p.Bet)
	if len(user.AvStrategy.GCYX.DMRecord) > 50 { //最近50局
		user.AvStrategy.GCYX.DMRecord = user.AvStrategy.GCYX.DMRecord[1:]
	}
	return nil
}

func (p *AviatorGCXYGCOverEvent) event(user *data.User) proto.Message {
	user.AvStrategy.GCYX.IsStateGC = false
	user.AvStrategy.GCYX.IsStateXZ = true
	user.AvStrategy.GCYX.EvoTimes = 0
	user.AvStrategy.GCYX.WinScore = 0
	user.AvStrategy.GCYX.AvgBet = 0
	return nil
}

func (p *AviatorGCXYXZOverEvent) event(user *data.User) proto.Message {
	if p.IsOver {
		user.AvStrategy.GCYX.IsStateXZ = false
		user.AvStrategy.GCYX.XZTimes = 0
	} else {
		user.AvStrategy.GCYX.XZTimes++
	}
	return nil
}

func (l *LHDStrategySettlementEvent) event(user *data.User) proto.Message {
	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		return nil
	}

	ctype := handler.GetChargeType(user)
	switch l.Id {
	case 1:
		user.LHDStrategy.XXSC.WinScore += l.WinScore
	case 2:
		// 求死不能
		user.LHDStrategy.QSBN.TriggerTimes++
	case 3:
		// 龙狂有祸
		if user.LHDStrategy.LKYH.BSRounds > 0 {
			if user.LHDStrategy.LKYH.BSRounds == strategy.LKYH.N[ctype] {
				user.LHDStrategy.LKYH.TriggerTimes++
				user.LHDStrategy.LKYH.TZ++
				user.LHDStrategy.LKYH.BSTimes++
			}

			user.LHDStrategy.LKYH.NZ++
			user.LHDStrategy.LKYH.BSRounds--

			if user.LHDStrategy.LKYH.BSRounds <= 0 {
				user.LHDStrategy.LKYH.BSRounds = 0
				// 进cd
				user.LHDStrategy.LKYH.BSCoolDown = utils.BsonNow().Unix() + strategy.LKYH.C[ctype]
			}
		} else {
			user.LHDStrategy.LKYH.Suppress++
		}
	case 4:
		user.LHDStrategy.GCYX.R++
		if l.WinScore > 0 {
			user.LHDStrategy.GCYX.Win += l.WinScore
		} else {
			user.LHDStrategy.GCYX.Lose += l.WinScore
		}
	default:
		user.LHDStrategy.XXSC.WinLength = 0
	}

	if user.LHDStrategy.XXSC.DisturbCoolDown > 0 {
		user.LHDStrategy.XXSC.DisturbCoolDown--
	}

	return nil
}

func (l *LHDXXSCSettlementEvent) event(user *data.User) proto.Message {
	s := config.GetLHDStrategy()
	if s.Id == 0 {
		return nil
	}

	ctype := handler.GetChargeType(user)
	// if l.Disturb {
	// 	user.LHDStrategy.XXSC.WinLength = 0
	// 	user.LHDStrategy.XXSC.DisturbCoolDown = s.XXSC.B[ctype]
	// } else {
	// 	user.LHDStrategy.XXSC.WinLength++
	// }
	if user.Diamond >= s.XXSC.M[ctype] {
		user.LHDStrategy.XXSC.TriggerTimes++
		user.LHDStrategy.XXSC.MaxTriggerTimes++
	}

	user.LHDStrategy.XXSC.Evo = true
	return nil
}

func (l *SevenStrategySettlementEvent) event(user *data.User) proto.Message {
	strategy := config.GetSevenStrategy()
	if strategy.Id == 0 {
		return nil
	}

	ctype := handler.GetChargeType(user)
	switch l.Id {
	case 1:
		user.SevenStrategy.XXSC.WinScore += l.WinScore
	case 2:
		// 求死不能
		user.SevenStrategy.QSBN.TriggerTimes++
	case 3:
		// 7管严
		if user.SevenStrategy.LKYH.BSRounds > 0 {
			if user.SevenStrategy.LKYH.BSRounds == strategy.QGY.N[ctype] {
				user.SevenStrategy.LKYH.TriggerTimes++
				user.SevenStrategy.LKYH.TZ++
				user.SevenStrategy.LKYH.BSTimes++
			}

			user.SevenStrategy.LKYH.NZ++
			user.SevenStrategy.LKYH.BSRounds--

			if user.SevenStrategy.LKYH.BSRounds <= 0 {
				user.SevenStrategy.LKYH.BSRounds = 0
				// 进cd
				user.SevenStrategy.LKYH.BSCoolDown = utils.BsonNow().Unix() + strategy.QGY.C[ctype]
			}
		} else {
			user.SevenStrategy.LKYH.Suppress++
		}
	case 4:
		user.SevenStrategy.GCYX.R++
		if l.WinScore > 0 {
			user.SevenStrategy.GCYX.Win += l.WinScore
		} else {
			user.SevenStrategy.GCYX.Lose += l.WinScore
		}
	default:
		user.SevenStrategy.XXSC.WinLength = 0
	}

	if user.SevenStrategy.XXSC.DisturbCoolDown > 0 {
		user.SevenStrategy.XXSC.DisturbCoolDown--
	}

	return nil
}

func (l *SevenXXSCSettlementEvent) event(user *data.User) proto.Message {
	s := config.GetSevenStrategy()
	if s.Id == 0 {
		return nil
	}

	ctype := handler.GetChargeType(user)
	// if l.Disturb {
	// 	user.SevenStrategy.XXSC.WinLength = 0
	// 	user.SevenStrategy.XXSC.DisturbCoolDown = s.XXSC.B[ctype]
	// } else {
	// 	user.SevenStrategy.XXSC.WinLength++
	// }
	if user.Diamond >= s.XXSC.M[ctype] {
		user.SevenStrategy.XXSC.TriggerTimes++
		user.SevenStrategy.XXSC.MaxTriggerTimes++
	}

	user.SevenStrategy.XXSC.Evo = true
	return nil
}

// vip bank 游戏任务事件触发
func (p *VBGameTaskEvent) event(user *data.User) proto.Message {
	// 未充值转正不开启活动
	if user.GetMoney() <= 0 {
		return nil
	}
	var changed bool
	for _, task := range user.VBTask {
		if task.Prize != 1 {
			continue
		}
		for _, tt := range task.TaskTypes {
			_ = tt
			typeRecord := table.GetTables().VBGameTaskTypeTable.Get(tt.Id)
			if typeRecord == nil {
				glog.Errorf("vb task type not exists %d", tt.Id)
				continue
			}
			progress := tt.Progress
			// 判断是否触发任务
			switch typeRecord.TaskType {
			case 1: // 游戏局数
				if utils.SliceIn(p.GameType, typeRecord.Gtype...) {
					tt.Progress++
				}
			case 2: // 游戏赢局数
				if p.Win && utils.SliceIn(p.GameType, typeRecord.Gtype...) {
					tt.Progress++
				}
			case 101: // tp,ak47,joker特定牌型获胜
				if utils.SliceIn(p.GameType, typeRecord.Gtype...) &&
					utils.SliceIn(p.GameType, int32(pb.HUA), int32(pb.HUA2), int32(pb.AK47), int32(pb.JOKER)) {
					if p.Win && p.HuaType == uint32(typeRecord.Arg2) {
						tt.Progress++
					}
				}
			case 102: // rummy在n回合获胜m局
				if utils.SliceIn(p.GameType, int32(pb.RUMMY), int32(pb.RUMMY2)) {
					if p.Win && p.ActTimes <= typeRecord.Arg2 {
						tt.Progress++
					}
				}
			case 103: // crash aviator在n倍以上逃脱n次
				if utils.SliceIn(p.GameType, typeRecord.Gtype...) &&
					utils.SliceIn(p.GameType, int32(pb.CRASH), int32(pb.PLANE)) {
					glog.Infof("crash event: %v, %v", p, typeRecord)
					if !p.Win {
						continue
					}
					multi, err := strconv.ParseFloat(typeRecord.Arg3, 64)
					if err != nil {
						glog.Errorf("vb task crash multi parse error: %s, %v", typeRecord.Arg3, err)
						continue
					}
					if p.CrashEscape >= multi {
						tt.Progress++
					}
				}
			default:
				glog.Errorf("unsupport task type %d: %v", typeRecord.TaskType, p)
				return nil
			}

			// 任务进度更新
			if progress != tt.Progress {
				changed = true

				// 获取最大进度
				maxProgress := handler.GetVBTaskTypeMaxProgress(typeRecord)
				if tt.Progress >= maxProgress {
					task.Prize = 2 // 待领奖
				}
			}
		}
	}

	if changed {
		// 任务更新消息
		msg := handler.BuildVBTaskDataMsg(user)
		return &pb.ActivityDataNtf{VbTask: msg}
	}
	return nil
}

func (p *CrashJackpotEvent) event(user *data.User) proto.Message {
	if p.Jackpot <= 0 {
		return nil
	}

	user.CrashStrategy.JCFK.JackpotVal += p.Jackpot
	return nil
}

func (p *CrashGCYXTriggerEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.GCYX.IsStateGC = true
	user.CrashStrategy.GCYX.M++
	user.CrashStrategy.GCYX.DailyGCTimes++
	user.CrashStrategy.GCYX.AvgBet = p.AvgBet
	return nil
}

func (p *CrashGCYXScoreEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.GCYX.R++
	user.CrashStrategy.GCYX.WinScore += p.Score
	return nil
}

func (p *CrashGCYXBetEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.GCYX.DMRecord = append(user.CrashStrategy.GCYX.DMRecord, p.Bet)
	if len(user.CrashStrategy.GCYX.DMRecord) > 50 { //最近50局
		user.CrashStrategy.GCYX.DMRecord = user.CrashStrategy.GCYX.DMRecord[1:]
	}
	return nil
}

func (p *CrashGCXYGCOverEvent) event(user *data.User) proto.Message {
	user.CrashStrategy.GCYX.IsStateGC = false
	user.CrashStrategy.GCYX.IsStateXZ = true
	user.CrashStrategy.GCYX.EvoTimes = 0
	user.CrashStrategy.GCYX.WinScore = 0
	user.CrashStrategy.GCYX.AvgBet = 0
	return nil
}

func (p *CrashGCXYXZOverEvent) event(user *data.User) proto.Message {
	if p.IsOver {
		user.CrashStrategy.GCYX.IsStateXZ = false
		user.CrashStrategy.GCYX.XZTimes = 0
	} else {
		user.CrashStrategy.GCYX.XZTimes++
	}
	return nil
}

func (p *LHDLKYHBSEvent) event(user *data.User) proto.Message {
	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		return nil
	}

	ctype := handler.GetChargeType(user)

	user.LHDStrategy.LKYH.BSRounds = strategy.LKYH.N[ctype]
	user.LHDStrategy.LKYH.BSCoolDown = utils.BsonNow().Unix() + strategy.LKYH.C[ctype]
	user.LHDStrategy.LKYH.BSTimes++

	return nil
}

func (p *LHDLKYHTriggerEvent) event(user *data.User) proto.Message {
	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		return nil
	}

	user.LHDStrategy.LKYH.RZ++

	return nil
}

func (s *LHDStrategyResetEvent) event(user *data.User) proto.Message {
	for _, s := range s.Strategys {
		switch s {
		case 3:
			// 龙狂有祸
			user.LHDStrategy.LKYH.TZ = 0
			user.LHDStrategy.LKYH.TriggerTimes = 0
			user.LHDStrategy.LKYH.BSCoolDown = 0
			user.LHDStrategy.LKYH.BSRounds = 0
		}
	}

	return nil
}

func (s *ABStrategySettlementEvent) event(user *data.User) proto.Message {
	strategy := config.GetABStrategy()
	if strategy.Id == 0 {
		return nil
	}

	switch s.Id {
	case 1:
		// 安能求死
		user.ABStrategy.ANQS.TriggerTimes++
	case 2:
		// 安然躺赢
		ctype := handler.GetChargeType(user)
		if user.Diamond >= strategy.ARTY.M[ctype] {
			user.ABStrategy.ARTY.U++
			user.ABStrategy.ARTY.WinScore = 0
		}
		user.ABStrategy.ARTY.EvoTimes = true
		if s.WinScore > 0 {
			user.ABStrategy.ARTY.WinScore += s.WinScore
		}
	default:

	}

	return nil
}

func (s *CPStrategySettlementEvent) event(user *data.User) proto.Message {
	strategy := config.GetCPStrategy()
	if strategy.Id == 0 {
		return nil
	}

	switch s.Id {
	case 2:
		// 来易去难
		user.CPStrategy.LYQN.TriggerTimes++
	case 1:
		// 来玩就赢
		ctype := handler.GetChargeType(user)
		if user.Diamond >= strategy.LWJY.M[ctype] {
			user.CPStrategy.LWJY.U++
			user.CPStrategy.LWJY.UZ++
			user.CPStrategy.LWJY.WinScore = 0
		}
		user.CPStrategy.LWJY.EvoTimes = true
		if s.WinScore > 0 {
			user.CPStrategy.LWJY.WinScore += s.WinScore
		}
	}

	return nil
}

func (p *SevenQGYTriggerEvent) event(user *data.User) proto.Message {
	strategy := config.GetSevenStrategy()
	if strategy.Id == 0 {
		return nil
	}

	user.SevenStrategy.LKYH.RZ++

	return nil
}

func (s *RBStrategySettlementEvent) event(user *data.User) proto.Message {
	strategy := config.GetRBStrategy()
	if strategy.Id == 0 {
		return nil
	}

	switch s.Id {
	case 1:
		// 红运当头
		ctype := handler.GetChargeType(user)
		if user.Diamond >= strategy.HYDT.M[ctype] {
			user.RBStrategy.HYDT.U++
			user.RBStrategy.HYDT.WinScore = 0
		}
		user.RBStrategy.HYDT.EvoTimes = true
		if s.WinScore > 0 {
			user.RBStrategy.HYDT.WinScore += s.WinScore
		}
	case 2:
		// 绝处逢生
		user.RBStrategy.JCFS.TriggerTimes++
	}

	return nil
}

func (b *BreakingGiftEvent) event(user *data.User) proto.Message {
	if user.GiftPopMap == nil {
		user.GiftPopMap = make(map[string]data.GiftPop)
	}

	pop := false
	nowTime := utils.BsonNow().Unix()

	bean := table.GetTables().GiftPopTable.Get("tcpc1")

	if bean.CarryAmount > 0 && int64(bean.CarryAmount) <= user.Diamond {
		return nil
	}
	if p, ok := user.GiftPopMap["tcpc1"]; ok {
		if bean.LimitedTime > 0 && nowTime-p.EntryTime >= int64(bean.LimitedTime*60) {
			// 过期了
			return nil
		}
		if p.TodayPop < bean.DailyPop && nowTime-p.LastPopTime > int64(bean.CoolDown*60) {
			pop = true
			p.LastPopTime = nowTime
			p.TodayPop++
			user.GiftPopMap["tcpc1"] = p
		}
	} else {
		pop = true
		user.GiftPopMap["tcpc1"] = data.GiftPop{
			Id:          "tcpc1",
			LastPopTime: nowTime,
			EntryTime:   nowTime,
			TodayPop:    1,
		}
	}

	if pop {
		ntf := &pb.GiftWelfareNtf{Gtype: 2}
		gifts := make([]*pb.GiftWelfare, 0)
		for _, g := range bean.GiftId {
			gift := table.GetTables().GiftRechargeTable.Get(g)
			if gift == nil || gift.GiftType != 4 || gift.Show[user.RegistArea] == 0 {
				continue
			}
			var giveRatio int32
			for _, v := range gift.GiveRatio[user.RegistArea].Nums {
				giveRatio += v
			}
			b := &pb.GiftWelfare{
				Id:     gift.Id,
				Price:  gift.Price,
				Give:   giveRatio,
				Weight: gift.ShowWeight[user.RegistArea],
			}
			gifts = append(gifts, b)
		}
		sort.Slice(gifts, func(i, j int) bool {
			return gifts[i].Weight > gifts[j].Weight
		})
		ntf.Gifts = gifts
		return ntf
	}
	return nil
}
