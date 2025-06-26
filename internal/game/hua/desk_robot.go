package hua

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"math"
)

const (
	ActionNull     int32 = iota
	ActionSee            //看牌
	ActionPack           //弃牌
	ActionCall           //跟注
	ActionRaise          //加注
	ActionBi             //发起比牌
	ActionReplyBi        //回复比牌
	ActionAgreeBi        //同意比牌
	ActionRejectBi       //拒绝比牌
)

var huaTypeIndexs = map[int]int{
	10: 11,
	11: 10,
	20: 9,
	21: 8,
	30: 7,
	31: 6,
	40: 5,
	41: 4,
	50: 3,
	51: 2,
	60: 1,
	61: 0,
}

var roomAnteIndexs = map[uint32]int{
	10:    0,
	100:   1,
	300:   2,
	500:   3,
	1000:  4,
	5000:  5,
	10000: 6,
}

func (t *Desk) getRobotExpressRecord() *tb.TpTpRobotExpressRecord {
	player := t.getPlayer(t.GetOnlyOnePlayer())
	if player != nil {
		chargeType := handler.GetChargeType(player)
		for _, express := range table.GetTables().TpRobotExpressTable.GetDataList() {
			for i, on := range express.UserChargeTypes {
				if on == 1 && chargeType == i {
					return express
				}
			}
		}
	}
	return table.GetTables().TpRobotExpressTable.Get("1001")
}

func (t *Desk) winRateChangePlayerCards() {
	exp := t.getRobotExpressRecord()
	if len(exp.WinRateEffect) == 0 {
		return
	}
	player := t.getPlayer(t.GetOnlyOnePlayer())
	if player == nil {
		return
	}
	carry := int32(player.GetScore())
	for _, effect := range exp.WinRateEffect {
		if effect.CarryMin <= carry && (effect.CarryMax < 0 || effect.CarryMax > carry) {
			t.winRateChangePlayerCards0(effect.Rate)
			return
		}
	}

}

// 胜率随中了换牌
func (t *Desk) winRateChangePlayerCards0(winRate int32) {
	if !utils.RandWan(int32(math.Abs(float64(winRate)))) {
		return
	}

	var maxSeat, secondSeat *data.DeskSeat
	var playerSeat *data.DeskSeat
	var robotSeats []*data.DeskSeat
	for _, seat := range t.seats {
		if t.getPlayer(seat.Userid).Robot {
			robotSeats = append(robotSeats, seat)
		} else {
			playerSeat = seat
		}
		if maxSeat == nil {
			maxSeat = seat
			continue
		}
		if algo.HuaCompare(seat.Cards, maxSeat.Cards) {
			secondSeat = maxSeat
			maxSeat = seat
		} else if secondSeat == nil {
			secondSeat = seat
		}
	}

	isRobotMax := t.getPlayer(maxSeat.Userid).Robot // 最大牌是否人机
	if winRate > 0 && isRobotMax {
		// 将玩家牌换到最大
		maxSeat.Cards, playerSeat.Cards = playerSeat.Cards, maxSeat.Cards
		// todo 记录干预
		zlog.Infof("%s win_rate_effect bigger: %d", t.GameId, winRate)
	}
	if winRate < 0 && !isRobotMax {
		// 将玩家牌换小
		i := utils.RandIntN(len(robotSeats))
		maxSeat.Cards, robotSeats[i].Cards = robotSeats[i].Cards, maxSeat.Cards
		zlog.Infof("%s win_rate_effect minimum: %d", t.GameId, winRate)
	}
}

// 人机策略初始化
func (t *Desk) robotStartInit() {
	// 找牌型最大的人机
	playerid := t.GetOnlyOnePlayer()
	playerSeat := t.getSeat(t.getSeatid(playerid))
	t.isPlayerHandMax = true

	// 人机类型
	var robotTypeChoices []utils.Choice
	rExp := t.getRobotExpressRecord()
	for t, rate := range rExp.RobotTypeRates {
		robotTypeChoices = append(robotTypeChoices, utils.Choice{Weight: int(rate), Item: t})
	}

	for _, seat := range t.seats {
		if role := t.getRole(seat.Userid); role == nil || !role.Robot {
			continue
		}
		// stake概率和极限倍数
		seat.StakeRateYi, seat.MaxMultiple = t.getCardsStakeRateAndMaxMultiple(seat.Cards)

		if playerSeat != nil {
			seat.BiggerThanPlayer = algo.HuaCompare(seat.Cards, playerSeat.Cards)
			if seat.BiggerThanPlayer {
				t.isPlayerHandMax = false
			}
		}
		// 人机风格类型, 人机在同一张桌子风格类型不变, 不重新随机
		var changeRobotType = seat.RobotType == 0
		// 人机风格切换频率
		if !changeRobotType && seat.SeatRounds/int32(seat.RobotTypeChangeTimes+1) >= t.getRobotExpressRecord().RobotTypeChangeRound {
			if utils.RandWan(t.getRobotExpressRecord().RobotTypeChangeRate) {
				changeRobotType = true
			}
		}
		if changeRobotType {
			seat.RobotTypeChangeTimes++
			c, err := utils.WeightedChoice(robotTypeChoices)
			if err != nil {
				glog.Errorf("robot type choice error: ", err)
			} else {
				seat.RobotType = c.Item.(int) + 1
			}

			if seat.RobotTypeChangeTimes > 1 {
				zlog.Infof("%s robot type change: %s, type->%d", t.GameId, seat.Userid, seat.RobotType)
			}
		}

		// 只有2人时增加一份看牌概率
		if len(t.roles) == 2 {
			seat.SeeRate += t.getRobotExpressRecord().SeeAddRates[seat.RobotType-1]
		}
		zlog.Infof("%s robot info %s, type=%d, stakeRate=%d, maxMultiple=%f", t.GameId, seat.Userid, seat.RobotType, seat.StakeRateYi, seat.MaxMultiple)
	}
}

// 当只剩两人时增加看牌概率
func (t *Desk) addRobotSeeRateWhenJust2Player() {
	alive, _, _ := t.getAliveNum()
	if alive == 2 {
		for seatid, seat := range t.seats {
			if t.isAlive(seatid) && t.isRobot(seatid) {
				seat.SeeRate += t.getRobotExpressRecord().SeeAddRates[seat.RobotType-1]
			}
		}
	}
}

// 弃牌或比牌输之后判断是否离开
func (t *Desk) robotLeaveWhenFinish(seatid uint32, win bool) {
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}
	if !t.isRobot(seatid) {
		return
	}

	// 只剩一个玩家了等比牌动画结束再离开
	alive, _, _ := t.getAliveNum()
	if alive == 1 {
		return
	}

	seat.TryLeave = true

	// 计算换桌概率
	if !seat.LeaveRateInited {
		seat.LeaveRate = t.getRobotExpressRecord().LevelDeskRate
		seat.LeaveRateInited = true
	}
	if win {
		seat.LeaveRate += t.getRobotExpressRecord().LevelDeskAddRate[1]
	} else {
		seat.LeaveRate += t.getRobotExpressRecord().LevelDeskAddRate[0]
	}

	var leave bool
	if seat.LeaveRate <= 0 {
		seat.LeaveRate = 0
	} else if seat.LeaveRate >= 10000 {
		seat.LeaveRate = 10000
		leave = true
	} else if utils.RandWan(seat.LeaveRate) {
		leave = true
	}

	var leaveMs int
	if leave {
		// 离开
		leaveMs = utils.RandMN(
			int(t.getRobotExpressRecord().LevelDeskDelay[0]),
			int(t.getRobotExpressRecord().LevelDeskDelay[1]))
		ntf := &pb.RobotLeaveNtf{RobotId: seat.Userid, LeaveMs: int32(leaveMs)}
		t.send2userid(seat.Userid, ntf)
		t.robotLeaving[seat.Userid] = true
	}

	zlog.Infof("%s robot leave %s: %v, rate=%d, ms=%d", t.GameId, seat.Userid, leave, seat.LeaveRate, leaveMs)
}

// 获取牌面stake概率和极限倍数
func (t *Desk) getCardsStakeRateAndMaxMultiple(handCards []uint32) (stakeRateYi int32, maxMultipleFixed float64) {
	cards := make([]uint32, len(handCards), len(handCards))
	copy(cards, handCards)
	algo.Sort(cards)
	indexs := algo.CardsToIndexs(cards)
	rvIndex := tb.TpRobotValueTableIndex{
		Card1: int32(indexs[0]),
		Card2: int32(indexs[1]),
		Card3: int32(indexs[2]),
	}
	robotValue := table.GetTables().TpRobotValueTable.GetByIndex(rvIndex)
	if robotValue == nil {
		glog.Errorf("not found robot value config: %v", cards)
		return
	}

	// 极限倍数修正系数
	var multipleRateFix float64 = 1.0
	index, ok := roomAnteIndexs[t.DeskData.Ante]
	if ok {
		multipleRateFix = t.getRobotExpressRecord().MultipleRateFix[index]
	}

	players := len(t.seats)
	var stakeRate string
	var maxMultiple float64
	switch players {
	default:
		glog.Errorf("unknown players: %d", players)
		return
	case 2:
		stakeRate = robotValue.StakeRate2
		maxMultiple = robotValue.MaxMultiple2
	case 3:
		stakeRate = robotValue.StakeRate3
		maxMultiple = robotValue.MaxMultiple3
	case 4:
		stakeRate = robotValue.StakeRate4
		maxMultiple = robotValue.MaxMultiple4
	case 5:
		stakeRate = robotValue.StakeRate5
		maxMultiple = robotValue.MaxMultiple5
	}
	// 百分比转亿
	var stakeRateF float64
	if _, err := fmt.Sscanf(stakeRate, "%f%%", &stakeRateF); err != nil {
		zlog.Errorf("stakeRate parse error: %s, %v, %#v", stakeRate, err, robotValue)
		return
	}
	stakeRateYi = int32(stakeRateF * 1000000.0)
	// 极限倍数
	maxMultipleFixed = maxMultiple * multipleRateFix
	return
}

// robotActState 人机操作行为
func (t *Desk) robotActState(msg *pb.JHPushActStateNtf) {
	userid := t.getUserid(msg.Seat)
	role := t.getRole(userid)
	if role == nil || !role.Robot {
		return
	}
	seatid := msg.Seat
	seat := t.getSeat(seatid)
	if seat == nil {
		return
	}

	rAction := new(pb.RAction)
	msg.RAction = rAction

	rExpress := t.getRobotExpressRecord()
	// 操作延时
	rAction.Delay = int32(utils.RandMN(int(rExpress.SelfActionDelay[0]), int(rExpress.SelfActionDelay[1])))

	var trace []int16
	defer func() {
		zlog.Infof("%s robot action %s action=%d(%d) see=%v(%d), trace=%v", t.GameId, seat.Userid, rAction.Action, rAction.Delay, rAction.ActionSee, rAction.SeeDelay, trace)
	}()

	// 回复比牌
	if msg.State&int32(pb.ACT_REPLY_BI) != 0 {
		rAction.Action = t.handleRobotReplyBi(seatid, seat, msg.NextSeat, &trace)
		return
	}

	var action int32
	var actionSee bool // 是否要看牌再操作
	var see = t.isSee(seatid)
	if !see {
		trace = append(trace, 41)
		seeRate := seat.SeeRate + rExpress.SeeRates[seat.RobotType-1] + rExpress.SeeAddRates[seat.RobotType-1]*t.getOtherSeeChaalTimes(seatid)
		if utils.RandWan(seeRate) {
			trace = append(trace, 42)
			actionSee = true
			see = true
		}
	}

	if !see { // 闷
		// blind概率随机2次
		blindRate := t.getRobotExpressRecord().BlindRates[seat.RobotType-1]
		// 乘doubleblind修正概率
		blindRate = int32(float64(blindRate) * t.getRobotExpressRecord().BlindRaiseStakeRateFix[seat.RobotType-1])
		if utils.RandWan(blindRate) && utils.RandWan(blindRate) {
			trace = append(trace, 13)
			rAction.Action = ActionRaise
		} else {
			trace = append(trace, 14)
			rAction.Action = ActionCall
		}
		return
	}

	// 判断stake概率
	isStake := utils.RandYi(seat.StakeRateYi)
	var logicSee bool
	if isStake {
		trace = append(trace, 15)
		action, logicSee = t.handleRobotStake0(seatid, seat, &trace)
	} else {
		trace = append(trace, 16)
		isMaxMultiple := t.isOverflowMultiple(seatid, seat, 1)
		zlog.Infof("%s robot pack %s, over=%v, see=%v, actAnte=%d, bet=%d, ante=%d, maxMultiple=%f", t.GameId, seat.Userid, isMaxMultiple, t.isSee(seatid), t.ActAnte, seat.Bet, t.DeskData.Ante, seat.MaxMultiple)
		action, logicSee = t.handleRobotPackV3(seatid, seat, isMaxMultiple, &trace)
	}
	actionSee = actionSee || logicSee

	// stake操作超过极限倍数
	if isStake && t.isActionOverflowMultiple(action, seatid, seat, actionSee || action == ActionBi) {
		trace = append(trace, 17)
		action, logicSee = t.handleRobotOverflowMultipeV3(seatid, seat, action, &trace)
	}
	actionSee = actionSee || logicSee

	// 比牌但是当前不能比
	if action == ActionBi && !(t.CanShow(t.DeskAct.ActState) || t.CanSideShow(t.DeskAct.ActState)) {
		if !t.isActionOverflowMultiple(ActionCall, seatid, seat, actionSee) {
			trace = append(trace, 19)
			action = ActionCall // 未到极限倍数下个注
		} else {
			trace = append(trace, 18)
			action = ActionPack // 到极限倍数了弃牌
		}

	}
	// 人机弃牌前看牌
	if action == ActionPack && !t.isSee(seatid) {
		trace = append(trace, 20)
		actionSee = true
	}

	// todo test 前5轮不走
	// if t.DeskAct.ActTimes < 5 && utils.SliceIn(action, ActionPack, ActionBi) {
	// 	action = ActionCall
	// }

	// 人机操作
	rAction.Action = action

	// 人机局内充值
	if utils.SliceIn(rAction.Action, ActionCall, ActionRaise, ActionBi) {
		// 动作是否触发局内充值
		rAction.ActionCharge = t.robotActionChargeInGame(seatid, seat, rAction.Action, actionSee)
	}

	// 人机看牌再操作
	if actionSee {
		trace = append(trace, 0)
		rAction.ActionSee = actionSee
		rAction.SeeDelay = int32(utils.RandMN(int(rExpress.SelfActionDelay[0]), int(rExpress.SelfActionDelay[1])))
	}
}

// 人机动作是否触发局内充值
func (t *Desk) robotActionChargeInGame(seatid uint32, seat *data.DeskSeat, action int32, actionBeforeSee bool) (charging bool) {
	// 充值过一次
	// if times, ok := t.ActRechargeTimes[seatid]; ok && times > 0 {
	// 	return
	// }

	player := t.getPlayer(seat.Userid)
	if player == nil {
		return
	}

	var actionCost int64
	switch action {
	case ActionCall, ActionBi:
		actionCost = t.ActAnte
		if t.isSee(seatid) || actionBeforeSee {
			actionCost *= 2
		}
	case ActionRaise:
		actionCost = t.ActAnte * 2
		if t.isSee(seatid) || actionBeforeSee {
			actionCost *= 2
		}
	}
	// 人机携带金额足够
	if actionCost <= player.GetScore() {
		return
	}

	// 下注超携带上限, 假装局内充值开始
	glog.Info("robot charge in game start: action cost=%d, score=%d", actionCost, player.GetScore())
	t.chargeInGameBegin(seatid)
	return true
}

// 计算人机局内充值金额
func (t *Desk) getRobotChargeInGameAmount(seatid uint32, seat *data.DeskSeat) int32 {
	conf := t.getRechargeAmountConfig(seatid)
	target := int32(t.getEstimateCash(seatid))

	for _, v := range conf.ValueRange {
		if v > target {
			return v
		}
	}
	// 最高档位
	return conf.ValueRange[len(conf.ValueRange)-1]
}

// 人机回复比牌逻辑
func (t *Desk) handleRobotReplyBi(seatid uint32, seat *data.DeskSeat, biSeat uint32, trace *[]int16) (action int32) {
	biPlayer := t.getPlayer(t.getUserid(biSeat))
	stakeAction := t.handleRobotStakeV1(seatid, seat, trace)
	stakeOverflowMultiple := t.isActionOverflowMultiple(stakeAction, seatid, seat, stakeAction == ActionBi)

	*trace = append(*trace, 4)
	// 人机比牌
	if biPlayer.Robot {
		// 此时stake是否触发极限倍数
		if stakeOverflowMultiple {
			*trace = append(*trace, 401)
			return ActionAgreeBi
		} else {
			if algo.HuaType(seat.Cards) >= algo.DuiZi { // 是否对子以上
				if utils.RandYi(seat.StakeRateYi) { // stake概率判断
					*trace = append(*trace, 402)
					return ActionRejectBi
				} else {
					*trace = append(*trace, 403)
					return ActionAgreeBi
				}
			} else {
				*trace = append(*trace, 404)
				return ActionAgreeBi
			}
		}
	}

	// 玩家比牌
	if stakeOverflowMultiple {
		*trace = append(*trace, 411)
		return ActionAgreeBi
	}
	if !(algo.HuaType(seat.Cards) >= algo.DuiZi) { // 是否对子以上
		*trace = append(*trace, 412)
		return ActionAgreeBi
	} else {
		if seat.BiggerThanPlayer { // 是否比玩家大
			*trace = append(*trace, 413)
			return ActionRejectBi
		} else {
			*trace = append(*trace, 414)
			return ActionAgreeBi
		}
	}
}

// 触发极限倍数后逻辑
func (t *Desk) handleRobotOverflowMultipe0(seatid uint32, seat *data.DeskSeat, stakeAction int32, trace *[]int16) (action int32, actionSee bool) {
	*trace = append(*trace, 3)
	if !(seat.BiggerThanPlayer || t.isPlayerPack()) { // 牌是否比玩家大
		*trace = append(*trace, 30)
		return t.handleRobotPack0(seatid, seat, true, trace)
	}
	// 是否还有其他机器人
	_, robot, _ := t.getAliveNum()
	if !(robot > 1) {
		// 取玩家牌型vh和极限倍数的高值判断是否触发极限倍数
		var multiple float64 = seat.MaxMultiple
		vh, _, _, ok := t.getPlayerVHRecord(t.GetOnlyOnePlayer())
		if ok {
			multiple = math.Max(multiple, vh)
			zlog.Infof("%s use multiple vh max: %f, %f", t.GameId, seat.MaxMultiple, vh)
		}
		if t.isOverflowMultipleWithValue(seatid, seat, multiple) {
			*trace = append(*trace, 31)
			return ActionBi, false
		} else {
			*trace = append(*trace, 35)
			return t.handleRobotStake0(seatid, seat, trace)
		}
	} else {
		if t.isBiggerRobot(seatid) { // 牌是否比其他机器人大
			*trace = append(*trace, 32)
			return stakeAction, false
		} else {
			if algo.HuaType(seat.Cards) >= algo.DuiZi { // 是否对子以上
				*trace = append(*trace, 33)
				// 是否有被拒绝的情况
				if seat.SideShowBeRejected {
					*trace = append(*trace, 36)
					return ActionPack, false
				} else {
					*trace = append(*trace, 37)
					return ActionBi, false
				}
			} else {
				*trace = append(*trace, 34)
				return ActionPack, false
			}
		}
	}
}

// 人机stake逻辑
func (t *Desk) handleRobotStake0(seatid uint32, seat *data.DeskSeat, trace *[]int16) (action int32, actionSee bool) {
	stakeRateFix := t.getRobotExpressRecord().RaiseStakeRateFix
	huaType := algo.HuaTypeUpOrDown(seat.Cards)
	rateFix := stakeRateFix[huaTypeIndexs[huaType]]

	// stake概率随机2次
	if utils.RandYi(int32(float64(seat.StakeRateYi)*rateFix)) && utils.RandYi(int32(float64(seat.StakeRateYi)*rateFix)) {
		*trace = append(*trace, 11)
		return ActionRaise, false
	} else {
		*trace = append(*trace, 12)
		return ActionCall, false
	}
}

// pack逻辑
func (t *Desk) handleRobotPack0(seatid uint32, seat *data.DeskSeat, isMaxMultiple bool, trace *[]int16) (action int32, actionSee bool) {
	*trace = append(*trace, 2)
	if algo.HuaType(seat.Cards) >= algo.DuiZi { // 是否对子以上
		if isMaxMultiple { // 是否极限倍数的pack
			*trace = append(*trace, 21)
			// 是否第一轮
			return t.handleRobotPackFirstRound0(seatid, seat, trace)
		} else {
			if !t.isBiggerRobot(seatid) { // 有比自己大的人机
				*trace = append(*trace, 22)
				return ActionBi, false
			} else {
				if seat.BiggerThanPlayer || t.isPlayerPack() { // 比玩家大
					*trace = append(*trace, 24)
					return t.handleRobotStake0(seatid, seat, trace)
				} else {
					*trace = append(*trace, 25)
					// 是否第一轮
					return t.handleRobotPackFirstRound0(seatid, seat, trace)
				}
			}
		}
	} else {
		*trace = append(*trace, 26)
		// 是否第一轮
		return t.handleRobotPackFirstRound0(seatid, seat, trace)
	}
}

// pack逻辑是否第一轮 ->
func (t *Desk) handleRobotPackFirstRound0(seatid uint32, seat *data.DeskSeat, trace *[]int16) (action int32, actionSee bool) {
	*trace = append(*trace, 201)
	if t.DeskAct.ActTimes == 0 { // 是否第一轮
		if !t.isBiggerCards(seatid, false) {
			*trace = append(*trace, 202)
			return ActionPack, false
		} else {
			// 本局赢后总返奖率是否大于等于风控基础返奖率
			if t.isPlayerWinOverRiskControlRate() {
				*trace = append(*trace, 203)
				return t.handleRobotStake0(seatid, seat, trace)
			} else {
				*trace = append(*trace, 204)
				return ActionPack, false
			}
		}
	} else {
		alive, _, _ := t.getAliveNum()
		if alive != 2 { // 当前是否只有2人
			if !t.isBiggerCards(seatid, false) { // 是否最大牌
				*trace = append(*trace, 205)
				return ActionPack, false
			} else {
				if t.isPlayerWinOverRiskControlRate() { // 赢后超过风控
					*trace = append(*trace, 206)
					return t.handleRobotStake0(seatid, seat, trace)
				} else {
					*trace = append(*trace, 207)
					return ActionPack, false
				}
			}
		} else {
			if !t.isBeforeChaal(seatid) { // 之前是否有过 chaal 和 doublechaal
				*trace = append(*trace, 208)
				if !t.isBiggerCards(seatid, false) { // 是否此时全场最大
					*trace = append(*trace, 209)
					return ActionPack, false
				} else {
					if !t.isPlayerWinOverRiskControlRate() { // 赢后超过风控
						*trace = append(*trace, 210)
						return ActionPack, false
					} else {
						// see 了直接 stake, 否则 see 了再 stake
						if t.isSee(seatid) {
							*trace = append(*trace, 211)
							return t.handleRobotStake0(seatid, seat, trace)
						} else {
							*trace = append(*trace, 212)
							action, _ = t.handleRobotStake0(seatid, seat, trace)
							actionSee = true
							return
						}
					}
				}
			} else {
				// 判断对手之前是否有过 chaal
				rivalSeatid, chaal := t.isNextPlayerBeforeChaal(seatid)
				if chaal {
					*trace = append(*trace, 213)
					return t.handleRobotPackFirstRoundRivalChaal0(seatid, seat, rivalSeatid, trace)
				} else {
					*trace = append(*trace, 214)
					return t.handleRobotPackFirstRoundRivalUnChaal0(seatid, seat, rivalSeatid, trace)
				}
			}
		}
	}
}

// pack逻辑非第一轮 对手chaal过
func (t *Desk) handleRobotPackFirstRoundRivalChaal0(seatid uint32, seat *data.DeskSeat, rivalSeatid uint32, trace *[]int16) (action int32, actionSee bool) {
	if t.isRobot(rivalSeatid) { // 对手非玩家
		if t.isOverflowMultiple(seatid, seat, 1) { // 是否触发极限倍数
			*trace = append(*trace, 2101)
			return ActionBi, false
		} else {
			*trace = append(*trace, 2102)
			return t.handleRobotStake0(seatid, seat, trace)
		}
	} else {
		rivalSeat := t.getSeat(rivalSeatid)
		if !algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 是否比对手大
			*trace = append(*trace, 2103)
			// 返奖率
			if t.isPlayerWinOverRiskControlRateWithAction(ActionBi, false, seatid) {
				*trace = append(*trace, 2104)
				return ActionPack, false
			} else {
				if !(algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp) { // 非大对子以上
					*trace = append(*trace, 2105)
					return ActionBi, false
				} else {
					// 牌型vh和极限倍数的低值判断是否触发极限倍数
					vh, _, _, _ := t.getPlayerVHRecord(t.GetOnlyOnePlayer())
					if t.isOverflowMultipleWithVH(seatid, seat, vh, false) {
						*trace = append(*trace, 2106)
						return ActionBi, false
					} else {
						*trace = append(*trace, 2107)
						return t.handleRobotStake0(seatid, seat, trace)
					}
				}
			}
		} else {
			// 玩家对于人机本次拿到的牌型是否有对应场次的vh记录
			vh, _, _, ok := t.getPlayerVHRecord(t.GetOnlyOnePlayer(), seat.Cards...)
			if !ok {
				if t.isOverflowMultiple(seatid, seat, 1) {
					*trace = append(*trace, 2108)
					return ActionBi, false
				} else {
					*trace = append(*trace, 2109)
					return t.handleRobotStake0(seatid, seat, trace)
				}
			} else {
				if vh > seat.MaxMultiple { // vh是否大于极限倍数
					if t.isOverflowMultipleWithVH(seatid, seat, vh, true) { // 用vh判断是否触发极限倍数
						*trace = append(*trace, 2110)
						return ActionBi, false
					} else {
						*trace = append(*trace, 2111)
						return t.handleRobotStake0(seatid, seat, trace)
					}
				} else {
					if t.isOverflowMultiple(seatid, seat, 1) { // 是否触发极限倍数
						*trace = append(*trace, 2112)
						return ActionBi, false
					} else {
						if t.isOverflowMultipleWithVH(seatid, seat, vh, true) { // 用vh判断是否触发极限倍数
							*trace = append(*trace, 2113)
							return ActionBi, false
						} else {
							*trace = append(*trace, 2114)
							return t.handleRobotStake0(seatid, seat, trace)
						}
					}
				}
			}
		}
	}
}

// pack逻辑非第一轮 对手未chaal过
func (t *Desk) handleRobotPackFirstRoundRivalUnChaal0(seatid uint32, seat *data.DeskSeat, rivalSeatid uint32, trace *[]int16) (action int32, actionSee bool) {
	rivalSeat := t.getSeat(rivalSeatid)
	if t.isRobot(rivalSeatid) { // 对手非玩家
		if t.isOverflowMultiple(seatid, seat, 1) { // 是否触发极限倍数
			*trace = append(*trace, 2201)
			return ActionBi, false
		} else {
			*trace = append(*trace, 2202)
			return t.handleRobotStake0(seatid, seat, trace)
		}
	} else {
		if !(algo.HuaTypeUpOrDown(seat.Cards) <= algo.DuiziUp) { // 非大对子以上
			if !algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 是否比对手大
				*trace = append(*trace, 2203)
				return ActionBi, false
			} else {
				if algo.HuaType(seat.Cards) == algo.GaoPai { // 是否是高牌
					*trace = append(*trace, 2204)
					return ActionBi, false
				} else {
					if !t.isPlayerWinOverRiskControlRate() { // 是否超过风控
						*trace = append(*trace, 2205)
						return ActionBi, false
					} else {
						if t.isOverflowMultiple(seatid, seat, 2) { // 是否触发两倍极限倍数
							*trace = append(*trace, 2206)
							return ActionBi, false
						} else {
							*trace = append(*trace, 2207)
							return t.handleRobotStake0(seatid, seat, trace)
						}
					}
				}
			}
		} else {
			if algo.HuaCompare(seat.Cards, rivalSeat.Cards) { // 比玩家大(此时2人对手为玩家)
				if t.isOverflowMultiple(seatid, seat, 2) { // 是否触发2倍极限倍数
					*trace = append(*trace, 2208)
					return ActionBi, false
				} else {
					*trace = append(*trace, 2209)
					return t.handleRobotStake0(seatid, seat, trace)
				}
			} else {
				if t.isOverflowMultiple(seatid, seat, 1) { // 已经触发极限倍数
					*trace = append(*trace, 2210)
					return ActionBi, false
				} else {
					if !t.isPlayerWinOverRiskControlRate() { // 玩家赢风控
						*trace = append(*trace, 2211)
						return t.handleRobotStake0(seatid, seat, trace)
					} else {
						// 超过极限倍数将人机牌换大stake
						overMax := t.isPlayerWinOverRiskControlMaxRate()
						var changedCard bool
						if overMax {
							*trace = append(*trace, 2212)
							// 如换牌失败, 上一层为否继续判断
							changedCard = t.handleChangeRobotBigger(seatid, seat)
						}
						if overMax && changedCard {
							*trace = append(*trace, 2213)
							return t.handleRobotStake0(seatid, seat, trace)
						} else {
							// 玩家是否有人机的牌的vh记录
							vh, _, _, ok := t.getPlayerVHRecord(t.GetOnlyOnePlayer(), seat.Cards...)
							if !ok {
								// 使用极限应变倍数判断是否触发极限倍数
								strainMaxMultiple := t.getStrainMaxMultiple(seat)
								if t.isOverflowMultipleWithValue(seatid, seat, strainMaxMultiple) {
									*trace = append(*trace, 2214)
									return ActionBi, false
								} else {
									*trace = append(*trace, 2215)
									return t.handleRobotStake0(seatid, seat, trace)
								}
							} else {
								if vh < t.getStrainMaxMultiple(seat) { // vh是否小于人机的应变极限倍数
									if t.isOverflowMultipleWithVH(seatid, seat, vh, true) { // vh判断是否触发极限倍数
										*trace = append(*trace, 2216)
										return ActionBi, false
									} else {
										*trace = append(*trace, 2217)
										return t.handleRobotStake0(seatid, seat, trace)
									}
								} else {
									// 使用极限应变倍数判断是否触发极限倍数
									strainMaxMultiple := t.getStrainMaxMultiple(seat)
									if t.isOverflowMultipleWithValue(seatid, seat, strainMaxMultiple) {
										*trace = append(*trace, 2218)
										return ActionBi, false
									} else {
										*trace = append(*trace, 2219)
										return t.handleRobotStake0(seatid, seat, trace)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// 获取应变极限倍数
func (t *Desk) getStrainMaxMultiple(seat *data.DeskSeat) float64 {
	huaType := algo.HuaTypeUpOrDown(seat.Cards)
	multipleRateStrain := t.getRobotExpressRecord().MultipleRateStrain[huaTypeIndexs[huaType]]
	return seat.MaxMultiple * multipleRateStrain
}

// 将人机的牌换成比玩家大一档的牌
func (t *Desk) handleChangeRobotBigger(seatid uint32, seat *data.DeskSeat) (ok bool) {
	playerid := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerid)
	playerSeat := t.getSeat(t.getSeatid(playerid))
	if player == nil || playerSeat == nil {
		return
	}
	playerHuaType := algo.HuaTypeUpOrDown(playerSeat.Cards)
	// 玩家3个A不换
	if playerHuaType == algo.BaoziUp && algo.Rank(playerSeat.Cards[0]) == algo.Ace {
		return
	}

	//将要换的手牌添加回牌堆
	deskCards := append(t.DeskGame.Cards, seat.Cards...)
	cards, remain := algo.GetCardHuaTypeBigger1(playerSeat.Cards, deskCards)
	if len(cards) > 0 {
		t.DeskGame.Cards = remain
		seat.Cards = cards
		ok = true

		// 更新stake概率和极限倍数
		seat.StakeRateYi, seat.MaxMultiple = t.getCardsStakeRateAndMaxMultiple(seat.Cards)
		// 更新手牌相关属性
		seat.BiggerThanPlayer = algo.HuaCompare(seat.Cards, playerSeat.Cards)
		if seat.BiggerThanPlayer {
			t.isPlayerHandMax = false
		}
	}
	return
}

// 人机stake逻辑
func (t *Desk) handleRobotStakeV1(seatid uint32, seat *data.DeskSeat, trace *[]int16) (action int32) {
	*trace = append(*trace, 1)
	stakeRateFix := t.getRobotExpressRecord().RaiseStakeRateFix
	huaType := algo.HuaTypeUpOrDown(seat.Cards)
	rateFix := stakeRateFix[huaTypeIndexs[huaType]]

	// stake概率随机2次
	if utils.RandYi(int32(float64(seat.StakeRateYi)*rateFix)) && utils.RandYi(int32(float64(seat.StakeRateYi)*rateFix)) {
		*trace = append(*trace, 11)
		return ActionRaise
	} else {
		*trace = append(*trace, 12)
		return ActionCall
	}
}

// 触发极限倍数后逻辑
func (t *Desk) handleRobotOverflowMultipeV1(seatid uint32, seat *data.DeskSeat, stakeAction int32, trace *[]int16) (action int32) {
	*trace = append(*trace, 3)
	if !(seat.BiggerThanPlayer || t.isPlayerPack()) { // 牌是否比玩家大
		*trace = append(*trace, 30)
		return t.handleRobotPackV1(seatid, seat, true, trace)
	}
	// 是否还有其他机器人
	_, robot, _ := t.getAliveNum()
	if !(robot > 1) {
		*trace = append(*trace, 31)
		return ActionBi
	} else {
		if t.isBiggerRobot(seatid) { // 牌是否比其他机器人大
			*trace = append(*trace, 32)
			return stakeAction
		} else {
			if algo.HuaType(seat.Cards) >= algo.DuiZi { // 是否对子以上
				*trace = append(*trace, 33)
				return ActionBi
			} else {
				*trace = append(*trace, 34)
				return ActionPack
			}
		}
	}
}

// pack逻辑
func (t *Desk) handleRobotPackV1(seatid uint32, seat *data.DeskSeat, isMaxMultiple bool, trace *[]int16) (action int32) {
	*trace = append(*trace, 2)
	if algo.HuaType(seat.Cards) >= algo.DuiZi { // 是否对子以上
		if isMaxMultiple { // 是否极限倍数的pack
			*trace = append(*trace, 21)
			// 是否第一轮
			return t.handleRobotPackFirstRoundV1(seatid, seat, trace)
		} else {
			if !t.isBiggerRobot(seatid) { // 有比自己大的人机
				*trace = append(*trace, 22)
				// return ActionBi
				// 走stake
				return t.handleRobotStakeV1(seatid, seat, trace)
			} else {
				if seat.BiggerThanPlayer || t.isPlayerPack() { // 比玩家大
					*trace = append(*trace, 24)
					// 改走 stake 流程
					return t.handleRobotStakeV1(seatid, seat, trace)
				} else {
					*trace = append(*trace, 25)
					// 是否第一轮
					return t.handleRobotPackFirstRoundV1(seatid, seat, trace)
				}
			}
		}
	} else {
		*trace = append(*trace, 26)
		// 是否第一轮
		return t.handleRobotPackFirstRoundV1(seatid, seat, trace)
	}
}

// pack逻辑是否第一轮 ->
func (t *Desk) handleRobotPackFirstRoundV1(seatid uint32, seat *data.DeskSeat, trace *[]int16) (action int32) {
	*trace = append(*trace, 201)
	if t.DeskAct.ActTimes == 0 { // 是否第一轮
		*trace = append(*trace, 202)
		return ActionPack
	} else {
		alive, _, _ := t.getAliveNum()
		if alive != 2 { // 当前是否只有2人
			*trace = append(*trace, 203)
			return ActionPack
		} else {
			if !t.isBeforeChaal(seatid) { // 判断之前是否chaal或doublechaal
				*trace = append(*trace, 204)
				return ActionPack
			} else {
				rivalSeatid, ok := t.isNextPlayerBeforeChaal(seatid)
				if !ok { // 判断对手之前是否chaal或doublechaal
					*trace = append(*trace, 205)
					return ActionBi
				} else {
					if algo.HuaCompare(seat.Cards, t.seats[rivalSeatid].Cards) { // 牌比对手大
						*trace = append(*trace, 206)
						return ActionBi
					} else {
						// show 是否触发极限倍数
						if t.isActionOverflowMultiple(ActionBi, seatid, seat, true) {
							*trace = append(*trace, 207)
							return ActionPack
						} else {
							*trace = append(*trace, 208)
							return ActionBi
						}
					}
				}
			}
		}
	}
}

// 获取玩家牌型的vh记录
// cards 不指定则为当前手牌
func (t *Desk) getPlayerVHRecord(playerId string, cards ...uint32) (vh float64, record [6]int64, key int32, ok bool) {
	playerSeat := t.getSeat(t.getSeatid(playerId))
	if playerSeat == nil {
		return
	}
	// huaType61(6bit) + ante(25bit[1kw]) = 32bit
	if len(cards) != 3 {
		cards = playerSeat.Cards
	}
	huaTypeUpDown := algo.HuaTypeUpOrDown(cards)
	key = (int32(huaTypeUpDown) << 25) | int32(t.DeskData.Ante)

	player := t.getPlayer(playerId)
	if player == nil || player.TpUserVHRecord == nil {
		return
	}
	record, ok = player.TpUserVHRecord[key]
	if ok {
		vh = float64(record[3]) / float64(record[1]) / float64(record[2])
	}
	return
}

// 玩家赢后是否大于等于风控基础返奖率, 基于人机极限倍数还会投注额
func (t *Desk) isPlayerWinOverRiskControlRate() bool {
	return t.isPlayerWinOverRiskControlRate0(t.getRobotExpressRecord().RiskControlRate)
}

// 玩家赢后是否大于等于风控极限返奖率, 基于人机极限倍数还会投注额
func (t *Desk) isPlayerWinOverRiskControlMaxRate() bool {
	return t.isPlayerWinOverRiskControlRate0(t.getRobotExpressRecord().RiskControlMaxRate)
}

func (t *Desk) isPlayerWinOverRiskControlRate0(riskControlRate float64) bool {
	playerid := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerid)
	if player == nil {
		return false
	}
	wins := player.CashOut + int32(player.Diamond) + int32(t.DeskGame.BetNum)
	// 人机还会投注额
	for seatid, seat := range t.seats {
		if !t.isRobot(seatid) || !t.isAlive(seatid) {
			continue
		}
		wins += int32(math.Max(0, seat.MaxMultiple*float64(t.DeskData.Ante)-float64(seat.Bet)))
	}
	var rechargeMoney = int32(player.Money) // 充值金额或默认
	if rechargeMoney <= 0 {
		rechargeMoney = t.getRobotExpressRecord().DefaultMoney
	}
	return float64(wins)/float64(rechargeMoney) >= riskControlRate
}

// 玩家赢后是否大于等于风控基础返奖率, 基于动作代价
func (t *Desk) isPlayerWinOverRiskControlRateWithAction(action int32, actionBeforeSee bool, actionSeatid uint32) bool {
	playerid := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerid)
	if player == nil {
		return false
	}

	var actionCost int64
	switch action {
	case ActionCall, ActionBi:
		actionCost = t.ActAnte
		if t.isSee(actionSeatid) || actionBeforeSee {
			actionCost *= 2
		}
	case ActionRaise:
		actionCost = t.ActAnte * 2
		if t.isSee(actionSeatid) || actionBeforeSee {
			actionCost *= 2
		}
	}
	wins := int32(actionCost) + player.CashOut + int32(player.Diamond) + int32(t.DeskGame.BetNum)
	var rechargeMoney = int32(player.Money) // 充值金额或默认
	if rechargeMoney <= 0 {
		rechargeMoney = t.getRobotExpressRecord().DefaultMoney
	}
	return float64(wins)/float64(rechargeMoney) >= t.getRobotExpressRecord().RiskControlRate
}

// 当前是否触发n倍极限倍数
func (t *Desk) isOverflowMultiple(seatid uint32, seat *data.DeskSeat, n float64) bool {
	if seat.MaxMultiple < 0 {
		return false
	}
	// 计算doublechaal成本
	raiseCost := t.ActAnte * 2
	if t.isSee(seatid) {
		raiseCost *= 2
	}
	return float64(raiseCost+seat.Bet)/float64(t.DeskData.Ante) >= seat.MaxMultiple*n
}

// 动作是否触发极限倍数
func (t *Desk) isActionOverflowMultiple(action int32, seatid uint32, seat *data.DeskSeat, actionBeforeSee bool) bool {
	if seat.MaxMultiple < 0 {
		return false
	}

	var actionCost int64
	switch action {
	case ActionCall, ActionBi:
		actionCost = t.ActAnte
		if t.isSee(seatid) || actionBeforeSee {
			actionCost *= 2
		}
	case ActionRaise:
		actionCost = t.ActAnte * 2
		if t.isSee(seatid) || actionBeforeSee {
			actionCost *= 2
		}
	}
	return float64(actionCost+seat.Bet)/float64(t.DeskData.Ante) >= seat.MaxMultiple
}

// 玩家的对应牌型vh和极限倍数的低值判断座位此极限倍数来判断是否触发极限倍数
func (t *Desk) isOverflowMultipleWithVH(seatid uint32, seat *data.DeskSeat, vh float64, forceVH bool) bool {
	multiple := seat.MaxMultiple
	if vh > 0 && (forceVH || vh < multiple) { // 没有vh时取极限倍数
		multiple = vh
		zlog.Infof("%s use multiple vh min: %f, %f", t.GameId, seat.MaxMultiple, vh)
	}
	if multiple < 0 {
		return false
	}
	// 计算doublechaal成本
	raiseCost := t.ActAnte * 2
	if t.isSee(seatid) {
		raiseCost *= 2
	}
	return float64(raiseCost+seat.Bet)/float64(t.DeskData.Ante) >= multiple
}

// 使用指定值判断是否触发极限倍数
func (t *Desk) isOverflowMultipleWithValue(seatid uint32, seat *data.DeskSeat, value float64) bool {
	if value < 0 {
		return false
	}
	// 计算doublechaal成本
	raiseCost := t.ActAnte * 2
	if t.isSee(seatid) {
		raiseCost *= 2
	}
	return float64(raiseCost+seat.Bet)/float64(t.DeskData.Ante) >= value
}

// 是否是当前场上最大人机
func (t *Desk) isBiggerRobot(curSeatid uint32) bool {
	return t.isBiggerCards(curSeatid, true)
}

// 是否是当前场上最大
// noPlayer 是否不算玩家
func (t *Desk) isBiggerCards(curSeatid uint32, noPlayer bool) bool {
	var biggerSeat uint32
	var biggerCard []uint32
	for seatid, seat := range t.seats {
		if !t.isAlive(seatid) || (noPlayer && !t.isRobot(seatid)) {
			continue
		}
		if biggerCard == nil {
			biggerCard = seat.Cards
			biggerSeat = seatid
		} else {
			if algo.HuaCompare(seat.Cards, biggerCard) {
				biggerSeat = seatid
				biggerCard = seat.Cards
			}
		}
	}
	return biggerSeat == curSeatid
}

// 计算其他玩家看牌后下注次数
func (t *Desk) getOtherSeeChaalTimes(curSeatid uint32) (chaalTimes int32) {
	for seatid := range t.seats {
		if seatid == curSeatid {
			continue
		}
		if !t.isAlive(seatid) {
			continue
		}

		status := t.getStatus(seatid)
		if status == nil {
			continue
		}
		for _, act := range status.ActHistory {
			if act&int32(pb.ACT_BLIND) == int32(pb.ACT_BLIND) { // 闷的不算
				continue
			}
			if act&int32(pb.ACT_CHAAL) == int32(pb.ACT_CHAAL) {
				chaalTimes++
			}
			if act&int32(pb.ACT_RAISE) == int32(pb.ACT_RAISE) {
				chaalTimes += 2
			}
		}
	}
	// 每一轮额外加一份看牌概率
	chaalTimes += t.DeskAct.ActTimes
	return
}

// 牌桌当前持牌未弃人数
func (t *Desk) getAliveNum() (alive, robot, player int) {
	for seatid, seat := range t.seats {
		if t.isAlive(seatid) {
			alive++
			if t.getPlayer(seat.Userid).Robot {
				robot++
			} else {
				player++
			}
		}
	}
	return
}

// 判断之前是否chaal或doublechaal
func (t *Desk) isBeforeChaal(seatid uint32) bool {
	status := t.getStatus(seatid)
	if status == nil {
		return false
	}
	return status.Chaal
}

// 判断对手之前是否有过chaal或者doublechaal
// return 对手seat,是否chaal
func (t *Desk) isNextPlayerBeforeChaal(curSeatid uint32) (seatid uint32, ok bool) {
	for seatid := range t.seats {
		if seatid != curSeatid && t.isAlive(seatid) {
			return seatid, t.isBeforeChaal(seatid)
		}
	}
	return 0, false
}

// 判断之前是否show被拒绝过
func (t *Desk) isBeforeSideShowBeRejected(seatid uint32, seat *data.DeskSeat) bool {
	return seat.SideShowBeRejected
}

// 玩家是否已弃牌
func (t *Desk) isPlayerPack() bool {
	playerid := t.GetOnlyOnePlayer()
	seatid := t.getSeatid(playerid)
	return !t.isAlive(seatid)
}

// 玩家chaal或doublechaal时人机看牌
func (t *Desk) robotSeeingInOther(curSeatid uint32) {
	for seatid, seat := range t.seats {
		if seatid == curSeatid ||
			!t.isRobot(seatid) ||
			t.isSee(seatid) ||
			!t.isAlive(seatid) {
			continue
		}

		seeRate := t.getRobotExpressRecord().OtherRoundSeeRate[seat.RobotType-1]
		if utils.RandWan(seeRate) {
			zlog.Infof("%s robot see in other %s ", t.GameId, seat.Userid)

			delay := t.getRobotExpressRecord().OtherSeeDelay
			seeDelay := int32(utils.RandMN(int(delay[0]), int(delay[1])))
			t.send2userid(seat.Userid, &pb.JHRobotSeeInOtherNtf{Seat: seatid, Delay: seeDelay})
		}
	}
}

// 记录玩家 偷鸡赢钱额/被偷鸡输钱额
func (t *Desk) hurtPlayerLog(win bool, playerId string, score int64) {
	player := t.getPlayer(playerId)
	var msg *pb.JHSetTpHurtScore
	if win && !t.isPlayerHandMax { // 偷鸡
		player.TpHurtWinRound++
		player.TpHurtWinScore += score
		msg = &pb.JHSetTpHurtScore{HurtScore: int32(score)}
		zlog.Infof("%s user hurt %s, %d", t.GameId, playerId, score)

	} else if !win && t.isPlayerHandMax { // 被偷鸡
		player.TpBeHurtLoseRound++
		player.TpBeHurtLoseScore += score
		msg = &pb.JHSetTpHurtScore{BeHurtScore: int32(score)}
		zlog.Infof("%s user be hurt %s, %d", t.GameId, playerId, score)
	}

	// 同步数据到网关
	if msg != nil {
		t.send2userid(playerId, msg)
	}
}

// 记录更新玩家vh
func (t *Desk) updatePlayerVHRecord(win bool, playerId string) {
	player := t.getPlayer(playerId)
	seat := t.getSeat(t.getSeatid(playerId))
	if player == nil || seat == nil {
		return
	}

	_, record, key, _ := t.getPlayerVHRecord(playerId)
	if player.TpUserVHRecord == nil {
		player.TpUserVHRecord = make(map[int32][6]int64)
	}
	// [牌型,底注,牌型次数,下注总金额,赢次数,输次数]
	record[0] = int64(key >> 25)
	record[1] = int64(t.DeskData.Ante)
	record[2]++
	record[3] += int64(seat.Bet)
	if win {
		record[4]++
	} else {
		record[5]++
	}
	player.TpUserVHRecord[key] = record

	// 同步数据到网关
	msg := &pb.JHVHRecordSync{
		VhKey:    key,
		VhRecord: record[:],
	}
	t.send2userid(playerId, msg)

	// vf记录
	vf := &pb.JHVFRecordSync{
		Userid:    player.Userid,
		Cards:     seat.Cards,
		BetAmount: int32(seat.Bet),
		Ante:      int32(t.DeskData.Ante),
	}
	myactor.Logger().Tell(vf)
}

// 游戏开始获取本局人机数量
func (t *Desk) calcRobotNumber() (robotNumber int) {
	// 获取空位数量
	playerNum, n := t.roleCountNum()
	robotNum := n - len(t.robotLeaving)
	emptySeatNum := int(t.Game.Count) - playerNum - robotNum

	var callRobot int
	for i := 0; i < emptySeatNum; i++ {
		rate := 10000 - (robotNum+i)*2500
		if rate > 0 && utils.RandWan(int32(rate)) {
			callRobot++
		} else {
			break
		}
	}
	robotNumber = robotNum + callRobot
	// 留个位置给玩家
	if robotNumber >= int(t.Game.Count) {
		robotNumber = int(t.Game.Count) - 1
	}
	zlog.Infof("%s robot number: %d=%d+%d, leaving=%d", t.GameId, robotNumber, robotNum, callRobot, len(t.robotLeaving))
	return
}
