package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"goserver/gen/tb"
	"math"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入或匹配桌子
func (rs *RoleActor) enterJH2MatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.hua_2").Name()
	msg.Gtype = int32(pb.HUA2) //金花
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

// 房间状态推送
func (r *RoleActor) recvJHPushStateNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHPushStateNtf)
	glog.Debugf("JHPushStateNtf %#v", msg)
	if msg.State == int32(pb.STATE_DEALING) { //发牌
		r.alive = true
		r.sendJHReady2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	} else if msg.State == int32(pb.STATE_OVER) { //游戏结束
		r.see = false
		r.jhstrategy = nil
		r.cards = []uint32{}
		r.state = 0
		r.core = false
		r.seeAndPack = false
		r.identity = ""
		r.afterRefusePack = false
	}
}

func (r *RoleActor) huaRunUniversal(msg *pb.JHPushActStateNtf) {
	action := r.huaUniversalAction(msg)
	action = r.huaRobotModelAction(action, msg)
	zlog.Infof("%s tp robot action=%d after model", r.Userid, action)
	r.huaRunAction(action)

	if action == ActionSee {
		action := r.huaUniversalAction(msg)
		r.huaRunAction(action)
	}
}

// huaRobotModelAction 新手、免费、剧情模式判断逻辑
func (r *RoleActor) huaRobotModelAction(action int, msg *pb.JHPushActStateNtf) int {
	model := msg.ActModel
	if model == nil {
		return action
	}
	zlog.Infof("%s tp robot action=%d, model=%d, %v", r.Userid, action, model.Model, model)
	if !model.PlayerAlive {
		return action
	}
	switch model.Model {
	case 0:
		return r.huaRobotModel0(action, model)
	// case 1: // 去掉免费模式逻辑
	// 	return r.huaRobotModel1(action, model, msg.IsNextPlayer)
	case 3:
		return r.huaRobotModel3(action, model, msg.IsPrevPlayer, msg.IsNextPlayer)
	}
	return action
}

// huaRobotModel3 剧情模式
func (r *RoleActor) huaRobotModel3(action int, model *pb.ActStateModel, isPrevPlayer, isNextPlayer bool) int {
	switch action {
	case ActionPack: // 弃牌逻辑
		return r.huaRobotModel3Pack(action, model, isPrevPlayer)
	case ActionAgreeBi: // 被比牌的时候
		fallthrough
	case ActionRejectBi:
		return r.huaRobotModel3ReplyBi(action, model, isNextPlayer)
	case ActionCall: // 跟注
		return r.huaRobotModel3Bet(action, model)
	case ActionRaise: // 加注
		return action
	case ActionBi: // 比牌
		return r.huaRobotModel3Bi(action, model)
	}
	return action
}

// huaRobotModel3Bet 跟注或者加注时
func (r *RoleActor) huaRobotModel3Bi(action int, model *pb.ActStateModel) int {
	if model.AliveNum > 3 {
		return ActionBi
	}
	// 玩家需跟注金额
	var playerCallBet int64 = model.ActAnte
	if model.PlayerSee {
		playerCallBet *= 2
	}
	// 玩家当前的金币数能否跟注/比牌，否比牌
	if model.PlayerDiamond < playerCallBet {
		return ActionBi
	}
	return ActionCall
}

// huaRobotModel3Bet 跟注或者加注时
func (r *RoleActor) huaRobotModel3Bet(action int, model *pb.ActStateModel) int {
	// 玩家局内充值了 -> 弃牌
	if (model.StoryPlus && model.PlayerChargeInGame > 1) || // 剧情局plus需充值两次
		(!model.StoryPlus && model.PlayerChargeInGame > 0) {
		return ActionPack
	}
	// 进入了加注超额阶段
	if model.SuperRaise {
		// 是否达到加注目标
		if model.PlayerDiamond < model.ActAnte {
			return ActionCall
		}
		// 前3大的牌，加注
		if model.HandCardNo > 0 && model.HandCardNo <= 3 {
			return ActionRaise
		}
		return ActionCall
	}
	return ActionCall
}

// huaRobotModel3ReplyBi 剧情模式被比牌
func (r *RoleActor) huaRobotModel3ReplyBi(action int, model *pb.ActStateModel, isNextPlayer bool) int {
	// 接受后是否游戏结束
	if model.AliveNum == 2 {
		return ActionAgreeBi
	}
	// 接受后参与人数是否为2
	if model.AliveNum != 3 {
		return ActionAgreeBi
	}

	// 玩家需跟注金额
	var playerCallBet int64 = model.ActAnte
	if model.PlayerSee {
		playerCallBet *= 2
	}

	// 下一个操作的是否是玩家 || 比牌是否玩家发起的，当前人机作为比牌发起者的上家
	if isNextPlayer {
		// 玩家是否进行了局内充值
		if (model.StoryPlus && model.PlayerChargeInGame > 1) || // 剧情局plus需充值两次
			(!model.StoryPlus && model.PlayerChargeInGame > 0) {
			return ActionAgreeBi
		}
		// 玩家是否有足够的金币跟注
		if model.PlayerDiamond >= playerCallBet {
			return ActionRejectBi
		}
		return ActionAgreeBi
	}

	// 玩家是否进行局内充值, 同意比牌
	if (model.StoryPlus && model.PlayerChargeInGame > 1) || // 剧情局plus需充值两次
		(!model.StoryPlus && model.PlayerChargeInGame > 0) {
		return ActionAgreeBi
	}
	// 玩家当前金币是否足够下轮操作, 否接受比牌
	if model.PlayerDiamond < playerCallBet {
		return ActionAgreeBi
	}
	return ActionRejectBi
}

// huaRobotModel3Pack 剧情模式弃牌
func (r *RoleActor) huaRobotModel3Pack(action int, model *pb.ActStateModel, isPrevPlayer bool) int {
	// 玩家是否进行了局内充值，弃牌
	if (model.StoryPlus && model.PlayerChargeInGame > 1) || // 剧情局plus需充值两次
		(!model.StoryPlus && model.PlayerChargeInGame > 0) {
		return action
	}
	var playerCallBet int64 = model.ActAnte
	if model.PlayerSee {
		playerCallBet *= 2
	}

	// 机器人弃牌后人数是否小于3
	if model.AliveNum <= 3 {
		// 玩家金币是否足够比牌，跟注
		if model.PlayerDiamond >= playerCallBet {
			if utils.RandWan(5000) { // 50%加注
				return ActionRaise
			}
			return ActionCall
		} else {
			// 当前桌内是否只有2人，跟注/弃牌
			if model.AliveNum == 2 {
				if utils.RandWan(5000) { // 50%加注
					return ActionRaise
				}
				return ActionCall
			} else {
				return ActionPack
			}
		}
	} else {
		// 进行比牌行为，是否是真人用户，否弃牌
		if !isPrevPlayer {
			return ActionPack
		}
		// 前3大的牌，否弃牌
		if !(model.HandCardNo > 0 && model.HandCardNo <= 3) {
			return ActionPack
		}
		// 玩家当前金币能否支撑玩家比牌，否比牌
		if model.PlayerDiamond < playerCallBet {
			return ActionBi
		}
		// 是否进入到加注超额注阶段
		if model.SuperRaise {
			return ActionRaise
		} else {
			return ActionCall
		}
	}
}

// huaRobotModel1 免费模式机器人
func (r *RoleActor) huaRobotModel1(action int, model *pb.ActStateModel, isNextPlayer bool) int {
	// playerWin := model.WinRateValid && model.WinRate > 0
	// 修正值为正则为增长方向
	playerWin := model.UserChangeCorrection > 0
	// 下一次玩家,当前人机下注额
	var playerNeedBet, robotBet int64
	var isBetAction bool
	switch action {
	case ActionCall:
		isBetAction = true
		playerNeedBet = model.ActAnte
		robotBet = model.ActAnte
	case ActionRaise:
		isBetAction = true
		playerNeedBet = model.ActAnte * 2
		robotBet = model.ActAnte * 2
	case ActionBi:
		isBetAction = true
		robotBet = model.ActAnte
	}
	if !isBetAction {
		return action
	}
	if playerWin {
		// 控制方向为玩家金币增长方向时候，机器人边界保护行为如下
		if model.PlayerDiamond+model.BetNum+robotBet < int64(math.Abs(float64(model.UserChangeCorrection))) {
			return action
		}
		// 是否8倍底注，弃牌
		if int64(model.Ante)*8 <= model.ActAnte {
			return ActionPack
		}
		// 是否最后一个人机，否弃牌
		if !(model.AliveNum == 2 && model.PlayerAlive) {
			return ActionPack
		}
		// 最后一个人机，比牌
		return ActionBi
	} else {
		// 控制方向为玩家金币减少方向的时候，机器人边界保护行为如下
		// 是否最后一个人机，否正常操作
		if !(model.AliveNum == 2 && model.PlayerAlive) {
			return action
		}
		// 玩家金币减少量是否大于等于档位目标
		if model.PlayerDiamond-playerNeedBet >= int64(math.Abs(float64(model.UserChangeCorrection))) {
			return ActionBi
		}
		return action
	}
}

// huaRobotModel0 新手保护模式机器人
func (r *RoleActor) huaRobotModel0(action int, model *pb.ActStateModel) int {
	playerWin := model.WinRateValid && model.WinRate > 0
	// 下一次玩家,人机下注额
	var playerNeedBet, robotBet int64
	var isBetAction bool
	switch action {
	case ActionCall:
		isBetAction = true
		playerNeedBet = model.ActAnte
		robotBet = model.ActAnte
	case ActionRaise:
		isBetAction = true
		playerNeedBet = model.ActAnte * 2
		robotBet = model.ActAnte * 2
	}
	if !isBetAction {
		return action
	}
	if model.PlayerSee {
		playerNeedBet = playerNeedBet * 2 //看牌double
	}
	if r.see {
		robotBet = robotBet * 2
	}

	if playerWin {
		// 加注跟注后用户是否跟得起
		// 用户跟不起，弃牌
		if playerNeedBet > model.PlayerDiamond {
			return ActionPack
		}
		// 玩家所赢金币是否超过了净增长的上线，弃牌
		if model.PlayerDiamond+model.BetNum+robotBet > int64(model.GrowthValueExceed) {
			return ActionPack
		}
		return action
	}

	// 用户不是最大的牌，加注跟注后用户是否会金币低于房间最低的金额, 弃牌
	if model.PlayerDiamond-playerNeedBet < int64(model.RoomMinAccess) {
		return ActionPack
	}
	return action
}

// 动作状态推送
func (r *RoleActor) recvJHPushActStateNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHPushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		// if r.core { //核心人机
		// 	r.huaRunCore(msg)
		// } else if r.seeAndPack { //看后即弃
		// 	r.huaRunSeeAndPack(msg)
		// } else if r.identity != "" { //主机僚机
		// 	r.huaRunIdentity(msg)
		// } else { //默认
		// 	r.huaRunDefalut(msg)
		// }
		r.huaRunUniversal(msg)
	} else { //其他人操作
		conf := r.getConf(msg.Turn)
		if conf != nil {
			if utils.RandWan(conf.SeeRate) {
				r.sendJHCoinSeeReq()
			}
		}
		// if r.alive && !r.see {
		// 	if r.core {
		// 		if utils.RandWan(3000) {
		// 			r.sendJHCoinSeeReq()
		// 		}
		// 	} else if r.seeAndPack {
		// 	} else if r.identity != "" {
		// 	} else {
		// 		if r.jhstrategy != nil {
		// 			if utils.RandWan(r.jhstrategy.SeeWeight[1]) {
		// 				r.sendJHCoinSeeReq()
		// 			}
		// 		}
		// 	}
		// }
	}
}

func (r *RoleActor) recvJHCoinReplyBiNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinReplyBiNtf)
	//自己发起比牌被拒（仅限玩家）
	if r.alive && !msg.Agree && msg.Seat == r.seat && !msg.Robot {
		conf := r.getConf(msg.Round)
		if conf != nil {
			if utils.RandWan(conf.AfterRefuse) {
				r.afterRefusePack = true
			}
		}
	}
}

// 人机策略推送
func (r *RoleActor) JHCoinRobotStrategyNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinRobotStrategyNtf)
	r.jhstrategy = msg
	r.core = msg.Core
	r.seeAndPack = msg.SeeAndPack
	r.identity = msg.Identity

	//局中换牌重置牌
	if msg.BetChange {
		r.cards = msg.ChangeCards
	}
	// glog.Infof("robot strategy %#v", msg)
}

// 游戏结束
func (r *RoleActor) JHCoinGameoverNtf2(ctx actor.Context) {
	if r.chargeInGame {
		// 假装局内充值了,回合结束离开
		glog.Infof("robot charge in game then leave: %v,%v,%v,%v", r.gtype, r.gameId, r.roomId, r.Userid)
		r.sendJHLeaveReq()
		return
	}
	if utils.RandWan(5000) {
		r.sendJHLeaveReq()
	}
}

// 人机策略信息推送
func (r *RoleActor) JHCoinRobotStrategyInfoNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinRobotStrategyInfoNtf)
	r.robotType = msg.RobotType
	r.count = msg.Count
	r.cardType = msg.CardType
	r.isMax = msg.IsMax
}

// 比牌结果推送
func (r *RoleActor) recvJHCoinBiNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinBiNtf)
	if r.alive && msg.Loseseat == r.seat {
		r.alive = false
	}

	//人机表情判断
	if msg.Winseat != r.seat && msg.Loseseat != r.seat { //没参与
		r.EmojiOtherBi(msg.Winseat, msg.Loseseat)
	} else if msg.Winseat == r.seat { //自己赢
		r.EmojiSelfBiWin(msg.Loseseat)
	}
}

func (r *RoleActor) recvJHCoinWaitTooLongNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHCoinWaitTooLongNtf)
	r.EmojiWaitLong(msg.Seat)
}

func (r *RoleActor) recvJHCoinEnterRoomRsp2(ctx actor.Context) {
	s2c := ctx.Message().(*pb.JHCoinEnterRoomRsp)
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		// 关闭节点
		glog.Errorf("enter hua fail user:%s,err:%v ", r.Userid, errcode)
		r.closeRs()
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	// r.rtype = roominfo.Rtype
	// r.dtype = roominfo.Dtype
	// r.roomid = roominfo.Roomid
	userinfo := s2c.GetUserinfo()

	for _, v := range userinfo {
		//只返回坐下玩家
		if v.Userid == r.Userid {
			glog.Debugf("comein user info -> %s", v.Userid)
			r.seat = v.Seat
			break
		} else { //记录其他玩家座位号
			r.seats = append(r.seats, v.Seat)
		}
	}

	//上桌发送表情
	r.EmojiDeskSend(r.GetRandSeat())
	// r.gameStart(s2c.Roominfo.State)
}

func (r *RoleActor) getConf(round int32) *tb.Tp2RobotStrategyRecord {
	round = round + 1
	if round > 5 {
		round = 5 //大于第五轮按第五轮处理
	}
	index := tb.Tp2RobotStrategyTableIndex{RobotType: r.robotType, Count: r.count, Round: round, CardType: r.cardType, IsMax: r.isMax}
	conf := table.GetTables().Tp2RobotStrategyTable.GetByIndex(index)
	if conf == nil { //如果找不到配置，按参与人数为2再查找
		index.Count = 2
		conf = table.GetTables().Tp2RobotStrategyTable.GetByIndex(index)
	}
	return conf
}

func (r *RoleActor) huaUniversalAction(msg *pb.JHPushActStateNtf) int {
	conf := r.getConf(msg.Turn)

	//处理回复比牌
	if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
		if conf == nil { //如果找不到配置，默认拒绝比牌
			return ActionRejectBi
		}
		if msg.IsNextPlayer { //比牌是真人
			agree := utils.RandWan(conf.AgreeByPlayer)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		} else { //比牌是机器
			agree := utils.RandWan(conf.AgreeByAi)
			if agree {
				return ActionAgreeBi
			} else {
				return ActionRejectBi
			}
		}
	}

	if conf == nil { //找不到配置，默认弃牌
		return ActionPack
	}

	if r.afterRefusePack { //被拒后弃牌（仅限玩家拒绝）
		return ActionPack
	}

	if !r.see { //看牌前
		var ch []utils.Choice
		ch = append(ch, utils.Choice{Weight: int(conf.SeeRate), Item: ActionSee})
		ch = append(ch, utils.Choice{Weight: int(conf.ChaalBLook), Item: ActionCall})
		ch = append(ch, utils.Choice{Weight: int(conf.ChaalBDLook), Item: ActionRaise})

		c, _ := utils.WeightedChoice(ch)
		return c.Item.(int)
	} else { //看牌后
		canBi := func() bool {
			if msg.State&int32(pb.ACT_SIDESHOW) != 0 || msg.State&int32(pb.ACT_SHOW) != 0 {
				return true
			} else {
				return false
			}
		}

		var ch []utils.Choice
		ch = append(ch, utils.Choice{Weight: int(conf.Chaal), Item: ActionCall})
		ch = append(ch, utils.Choice{Weight: int(conf.DoubleChaal), Item: ActionRaise})
		ch = append(ch, utils.Choice{Weight: int(conf.Pk), Item: ActionBi})
		ch = append(ch, utils.Choice{Weight: int(conf.GiveUp), Item: ActionPack})
		c, _ := utils.WeightedChoice(ch)

		if _, ok := c.Item.(int); !ok {
			glog.Debugf("无效动作 %v", conf)
		} else {
			glog.Debugf("有效动作 %v", conf)
		}

		action := c.Item.(int)
		if action == ActionBi && !canBi() {
			action = ActionCall
		}
		return action
	}
}
