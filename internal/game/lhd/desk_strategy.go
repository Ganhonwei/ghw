package lhd

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"sort"
)

func (t *Desk) selectStrategy() {
	// 心想事成
	t.triggerXXSC()
	// 求死不能
	t.triggerQSBN()
	// 龙狂有祸
	t.triggerLKYH()
	// 高潮涌现
	t.triggerGCYX()

	// 这个在最后面
	t.checkStrategy()
}

func (t *Desk) checkStrategy() {
	if len(t.LHDStrategys) <= 1 {
		return
	}

	sort.Slice(t.LHDStrategys, func(i, j int) bool {
		return t.LHDStrategys[i].Weight < t.LHDStrategys[j].Weight
	})

	strategy := config.GetLHDStrategy()

	newStrategy := make([]data.StrategyInfo, 0)
	mutexStrategy := make(map[int]string)

	for _, s := range t.LHDStrategys {
		if _, ok := mutexStrategy[s.Id]; ok {
			// 互斥了
			continue
		}
		switch s.Id {
		case XXSC:
			for _, v := range strategy.XXSC.Mutex {
				mutexStrategy[v] = ""
			}
		case QSBN:
			for _, v := range strategy.QSBN.Mutex {
				mutexStrategy[v] = ""
			}
		case LKYH:
			for _, v := range strategy.LKYH.Mutex {
				mutexStrategy[v] = ""
			}
		case GCYX:
			for _, v := range strategy.GCYX.Mutex {
				mutexStrategy[v] = ""
			}
		}
		newStrategy = append(newStrategy, s)
	}

	t.LHDStrategys = newStrategy
}

// 心想事成策略
func (t *Desk) triggerXXSC() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]

	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		glog.Error("no found xxsc strategy config, gameid:", t.GameId)
		return
	}

	if strategy.XXSC.BaseStrategy == nil || strategy.XXSC.O != 1 {
		return
	}

	ctype := handler.GetChargeType(role.User)
	if !handler.TiggerLHDStrategy(role.User, strategy.XXSC.BaseStrategy) || // 玩家类型是否满足
		role.Diamond+bet >= strategy.XXSC.M[ctype] ||
		role.LHDStrategy.XXSC.TriggerTimes >= strategy.XXSC.U[ctype] ||
		role.LHDStrategy.XXSC.MaxTriggerTimes >= strategy.XXSC.UZ[ctype] {
		return
	}

	info := data.StrategyInfo{Id: XXSC, Weight: strategy.XXSC.Weight}

	// 获取下注区域
	var seat int
	var maxBet int64
	noBetSeats := make([]int, 0) // 没下注的位置
	for s, bets := range t.LHSeatRoleBets {
		if b, ok := bets[userid]; ok {
			if s == Tie {
				continue
			}
			if b.GetSum() > maxBet {
				maxBet = b.GetSum()
				seat = int(s)
			}
		} else {
			noBetSeats = append(noBetSeats, int(s))
		}
	}

	if seat != int(Tie) && role.LHDStrategy.XXSC.WinScore < strategy.XXSC.C[ctype] {
		// 净赢分不能超过冷却线
		info.Seat = append(info.Seat, seat)
		// return
	}

	// 扰动
	// if role.LHDStrategy.XXSC.DisturbCoolDown <= 0 &&
	// 	role.LHDStrategy.XXSC.WinLength >= strategy.XXSC.A[ctype] &&
	// 	role.Diamond >= strategy.XXSC.D[ctype] {
	// 	// 开下的少的或者没下的位置
	// 	if len(noBetSeats) > 0 {
	// 		info.Seat = noBetSeats
	// 	} else {
	// 		// 开下得少的,除了和
	// 		info.Seat = make([]int, 0)
	// 		if t.LHSeatRoleBets[Dragon][userid].GetSum() > t.LHSeatRoleBets[Tiger][userid].GetSum() {
	// 			info.Seat = append(info.Seat, int(Tiger))
	// 		} else if t.LHSeatRoleBets[Dragon][userid].GetSum() < t.LHSeatRoleBets[Tiger][userid].GetSum() {
	// 			info.Seat = append(info.Seat, int(Dragon))
	// 		} else {
	// 			info.Seat = append(info.Seat, int(Tiger), int(Dragon))
	// 		}
	// 	}

	// 	info.Disturb = true
	// 	// role.LHDStrategy.XXSC.DisturbCoolDown = strategy.XXSC.B
	// }

	glog.Info("use XXSC strategy, gameid:", t.GameId)
	t.LHDStrategys = append(t.LHDStrategys, info)
	// if len(info.Seat) > 0 {
	// }
}

// 求死不能策略
func (t *Desk) triggerQSBN() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]

	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		glog.Error("no found xxsc strategy config, gameid:", t.GameId)
		return
	}

	if strategy.QSBN.BaseStrategy == nil || strategy.QSBN.O != 1 {
		return
	}

	// money := role.Money
	// if money <= 0 {
	// 	money = strategy.QSBN.
	// }

	ctype := handler.GetChargeType(role.User)
	if !handler.TiggerLHDStrategy(role.User, strategy.QSBN.BaseStrategy) || // 玩家类型是否满足
		role.Diamond >= strategy.QSBN.D[ctype] ||
		float64(int64(role.CashOut)+role.Diamond)/float64(role.Money) > strategy.QSBN.M[ctype] {
		return
	}

	if role.LHDStrategy.QSBN.TriggerTimes >= strategy.QSBN.T[ctype] && !utils.RandWan(strategy.QSBN.P[ctype]) {
		return
	}

	info := data.StrategyInfo{Id: QSBN, Weight: strategy.QSBN.Weight}

	for seat, bets := range t.LHSeatRoleBets {
		var winScore int64 = 0
		if b, ok := bets[userid]; ok {
			odds := t.LHDeskFree.OddsMap[seat] - 1
			winScore = b.GetSum() * int64(odds)
			if seat == Tie {
				winScore += (bet - b.GetSum()) / 2
			}
			if strategy.QSBN.RR[ctype] < float64(winScore+bet+role.Diamond+int64(role.CashOut))/float64(role.Money) {
				continue
			}

			if strategy.QSBN.PR[ctype] < (winScore + bet + role.Diamond + int64(role.CashOut) - int64(role.Money)) {
				continue
			}

			info.Seat = append(info.Seat, int(seat))
		}
	}

	if len(info.Seat) > 0 {
		glog.Infof("use QSBN strategy,seat:%v, gameid:%s", info.Seat, t.GameId)
		t.LHDStrategys = append(t.LHDStrategys, info)
	}
}

// 龙狂有祸策略
func (t *Desk) triggerLKYH() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]

	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		glog.Error("no found lkyh strategy config, gameid:", t.GameId)
		return
	}

	if strategy.LKYH.BaseStrategy == nil || strategy.LKYH.O != 1 {
		return
	}

	money := int64(role.Money)
	if money <= 0 {
		money = int64(strategy.LKYH.X)
	}

	ctype := handler.GetChargeType(role.User)

	kill := false
	// 是否处于生效状态
	if role.LHDStrategy.LKYH.BSRounds > 0 {
		// 执行倍杀
		kill = true
	}

	info := data.StrategyInfo{Id: LKYH, Weight: strategy.LKYH.Weight}

	WinScoreMap := make(map[uint32]int64)

	for i := 0; i < 3; i++ {
		if betmap, ok := t.LHSeatRoleBets[uint32(i)]; ok {
			if b, ok := betmap[userid]; ok {
				odds := t.LHDeskFree.OddsMap[uint32(i)] - 1
				winScore := b.GetSum() * int64(odds)
				if i == int(Tie) {
					winScore += (bet - b.GetSum()) / 2
				}

				WinScoreMap[uint32(i)] = winScore
				continue
			}
		}
		if i == int(Tie) {
			WinScoreMap[uint32(i)] = -bet / 2
		} else {
			WinScoreMap[uint32(i)] = -bet
		}
	}

	for seat, bets := range t.LHSeatRoleBets {
		var winScore int64 = 0
		if b, ok := bets[userid]; ok {
			odds := t.LHDeskFree.OddsMap[seat] - 1
			winScore = b.GetSum() * int64(odds)
			if seat == Tie {
				winScore += (bet - b.GetSum()) / 2
			}

			WinScoreMap[seat] = winScore
		}
	}

	var maxLose int64
	var seat []int
	for s, score := range WinScoreMap {
		if score > maxLose {
			continue
		} else if score == maxLose {
			seat = append(seat, int(s))
		} else {
			seat = []int{int(s)}
		}
		maxLose = score
	}

	canBS := true
	if !kill {
		if !handler.TiggerLHDStrategy(role.User, strategy.LKYH.BaseStrategy) {
			return
		}

		if role.LHDStrategy.LKYH.TriggerTimes >= strategy.LKYH.T[ctype] ||
			role.LHDStrategy.LKYH.TZ >= strategy.LKYH.TZ[ctype] {
			canBS = false
		}

		// 条件判断，任意一个条件满足都生效
		if float64(role.Diamond+int64(role.CashOut))/float64(money) < strategy.LKYH.PU[ctype] && // 条件1
			role.Diamond+int64(role.CashOut)-money < strategy.LKYH.PP[ctype] { // 条件2
			return
		}

		// 事件
		t.eventPost(userid, event.LHD_LKYH_Trigger, new(event.LHDLKYHTriggerEvent))
	}

	if !kill {
		// 先判断倍杀
		if len(role.LHDStrategy.RoundBet) > 0 && canBS {
			lastBet := role.LHDStrategy.RoundBet[len(role.LHDStrategy.RoundBet)-1]
			if bet >= 2*lastBet && utils.RandWan(strategy.LKYH.P4[ctype]) && utils.BsonNow().Unix() > role.LHDStrategy.LKYH.BSCoolDown {
				kill = true
				role.LHDStrategy.LKYH.BSRounds = strategy.LKYH.N[ctype]
				role.LHDStrategy.LKYH.BSCoolDown = utils.BsonNow().Unix() + strategy.LKYH.C[ctype]
				role.LHDStrategy.LKYH.BSTimes++
				// 事件
				t.eventPost(role.Userid, event.LHD_LKYH_BS, new(event.LHDLKYHBSEvent))
			}
		}

		if !kill && utils.RandWan(strategy.LKYH.P3[ctype]) {
			// 火力压制
			kill = true
			// 下了和就不触发
			for _, s := range seat {
				if s == int(Tie) {
					kill = false
					break
				}
			}
		}
	}

	if kill {
		info.Seat = seat
	}

	if len(info.Seat) > 0 {
		glog.Infof("use LKYH strategy,seat:%v, gameid:%s", info.Seat, t.GameId)
		t.LHDStrategys = append(t.LHDStrategys, info)
	}
}

func (t *Desk) triggerGCYX() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		glog.Error("no found gcyx strategy config, gameid:", t.GameId)
		return
	}

	if strategy.GCYX.BaseStrategy == nil || strategy.GCYX.O != 1 {
		return
	}
	// 玩家类型是否满足
	if !handler.TiggerLHDStrategy(role.User, strategy.GCYX.BaseStrategy) {
		return
	}

	defer func() {
		msg := &pb.LHGCYXSync{
			Highing:      role.LHDStrategy.GCYX.Highing,
			TriggerTimes: int32(role.LHDStrategy.GCYX.TriggerTimes),
			Hp:           int32(role.LHDStrategy.GCYX.Hp),
			C:            int32(role.LHDStrategy.GCYX.C),
			M:            int32(role.LHDStrategy.GCYX.M),
			BetAvg:       role.LHDStrategy.GCYX.BetAvg,
		}
		t.send2userid(userid, msg)
	}()

	ctype := handler.GetChargeType(role.User)

	if role.LHDStrategy.GCYX.Highing {
		// 判断是否结束高潮状态
		g := strategy.GCYX.G[ctype]
		a := float64(g) * role.LHDStrategy.GCYX.BetAvg
		cash := role.LHDStrategy.GCYX.Win + role.LHDStrategy.GCYX.Lose
		zlog.Infof("lhd gcyx in high cash=%d, a=%.2f, overHigh=%v, betAvg=%.6f", cash, a, float64(cash) >= a, role.LHDStrategy.GCYX.BetAvg)
		// 高潮盈利额大于等于高潮盈利限额a, 结束高潮状态
		if float64(cash) >= a || role.LHDStrategy.GCYX.BetAvg == 0.0 {
			role.LHDStrategy.GCYX.Highing = false
			role.LHDStrategy.GCYX.Hp = 0
			role.LHDStrategy.GCYX.C = strategy.GCYX.C[ctype]
			// 高潮结束输赢清零
			role.LHDStrategy.GCYX.Win = 0
			role.LHDStrategy.GCYX.Lose = 0
			return
		}

	} else {
		// 判断是否进入高潮状态
		if role.LHDStrategy.GCYX.C > 0 {
			role.LHDStrategy.GCYX.C--
			return
		}

		// 返奖率
		money := int64(role.Money)
		if money <= 0 {
			money = int64(strategy.GCYX.X)
		}
		rebateRate := float64(role.Diamond+int64(role.CashOut)) / float64(money)
		fp := strategy.GCYX.FP[ctype]
		zlog.Infof("lhd gcyx user %s, rebateRate=%.2f, fp=%.2f", role.Userid, rebateRate, fp)
		if rebateRate >= fp {
			return
		}
		// 超过单日高潮次数
		if role.LHDStrategy.GCYX.TriggerTimes > strategy.GCYX.S[ctype] {
			return
		}
		// 判断进入概率
		if role.LHDStrategy.GCYX.Hp == 0 {
			// 基础高潮率, 进入房间时hp清零
			role.LHDStrategy.GCYX.Hp = strategy.GCYX.HP[ctype]
		} else {
			// 递增高潮率
			role.LHDStrategy.GCYX.Hp += strategy.GCYX.HP1[ctype]
		}
		zlog.Infof("lhd gcyx user %s, hp=%d", role.Userid, role.LHDStrategy.GCYX.Hp)
		if !utils.RandWan(int32(role.LHDStrategy.GCYX.Hp)) {
			return
		}

		// 计算局均打码量
		f := strategy.GCYX.F[ctype]
		betLen := len(role.LHDStrategy.RoundBet)
		f = utils.Min(f, betLen)
		roundBets := role.LHDStrategy.RoundBet[betLen-f:]
		var betSum int64
		for _, b := range roundBets {
			betSum += b
		}
		// 还没有打码记录
		if betSum == 0 || f == 0 {
			return
		}

		zlog.Infof("lhd gcyx user %s in strategy", role.Userid)

		role.LHDStrategy.GCYX.BetAvg = float64(betSum) / float64(f)
		role.LHDStrategy.GCYX.Hp = 0
		role.LHDStrategy.GCYX.Highing = true
		role.LHDStrategy.GCYX.TriggerTimes++
		role.LHDStrategy.GCYX.M++
	}

	// 进入策略，找最高下注额位置
	info := data.StrategyInfo{Id: GCYX, Weight: strategy.GCYX.Weight}
	// 位置下注详情
	var maxSeat uint32
	var maxSeatBet int64
	for seat, bets := range t.LHSeatRoleBets {
		if b, ok := bets[userid]; ok {
			betSum := b.GetSum()
			if betSum > maxSeatBet {
				maxSeat = seat
				maxSeatBet = betSum
				continue
			}
		}
	}
	// 下注额最高是和走自然概率
	if maxSeat != Tie {
		info.Seat = []int{int(maxSeat)}
	}
	t.LHDStrategys = append(t.LHDStrategys, info)
}

// func (t *Desk) xxscTrigger(userid string) {
// 	user := t.roles[userid]
// 	if user == nil {
// 		return
// 	}

// 	if user.LHDStrategy.XXSC.DisturbCoolDown > 0 {
// 		user.LHDStrategy.XXSC.DisturbCoolDown--
// 	}
// }

// 策略结算
func (t *Desk) strategySettlement(userid string) {
	// t.xxscTrigger(userid)

	switch t.LHDeskFree.TriggerStrategy {
	case XXSC:
		t.XXSCSettlement(userid)
	case QSBN:
		t.QSBNSettlement(userid)
	case LKYH:
		t.LKYHSettlement(userid)
	case GCYX:
		t.GCYXSettlement(userid)
	}

	// 策略
	t.eventPost(userid, event.LHD_Strategy, &event.LHDStrategySettlementEvent{Id: t.LHDeskFree.TriggerStrategy, WinScore: t.ScoreMap[userid].GetSum()})
}

// 心想事成策略结算
func (t *Desk) XXSCSettlement(userid string) {
	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		return
	}

	user := t.roles[userid]
	if user == nil {
		return
	}

	ctype := handler.GetChargeType(user.User)

	for _, s := range t.LHDStrategys {
		if s.Id != XXSC {
			continue
		}
		// if s.Disturb {
		// 	user.LHDStrategy.XXSC.WinLength = 0
		// 	user.LHDStrategy.XXSC.DisturbCoolDown = strategy.XXSC.B[ctype]
		// } else {
		// 	user.LHDStrategy.XXSC.WinLength++
		// }

		//输赢分
		win := t.ScoreMap[userid]
		user.LHDStrategy.XXSC.WinScore += win.GetSum()

		if user.Diamond >= strategy.XXSC.M[ctype] {
			user.LHDStrategy.XXSC.TriggerTimes++
			user.LHDStrategy.XXSC.MaxTriggerTimes++
		}

		user.LHDStrategy.XXSC.Evo = true
		// 策略
		t.eventPost(userid, event.LHD_XXSC_Strategy, &event.LHDXXSCSettlementEvent{Disturb: s.Disturb})
	}
}

// 求死不能策略结算
func (t *Desk) QSBNSettlement(userid string) {
	strategy := config.GetLHDStrategy()
	if strategy.Id == 0 {
		return
	}

	if role, ok := t.roles[userid]; ok {
		role.LHDStrategy.QSBN.TriggerTimes++
	}
}

// 龙狂有祸结算
func (t *Desk) LKYHSettlement(userid string) {
	role := t.roles[userid]
	if role == nil {
		return
	}

	strategy := config.GetLHDStrategy()
	ctype := handler.GetChargeType(role.User)

	if role.LHDStrategy.LKYH.BSRounds > 0 {
		if role.LHDStrategy.LKYH.BSRounds == strategy.LKYH.N[ctype] {
			role.LHDStrategy.LKYH.TriggerTimes++
			role.LHDStrategy.LKYH.TZ++
		}

		role.LHDStrategy.LKYH.BSRounds--

		if role.LHDStrategy.LKYH.BSRounds <= 0 {
			role.LHDStrategy.LKYH.BSRounds = 0
			// 进cd
			role.LHDStrategy.LKYH.BSCoolDown = utils.BsonNow().Unix() + strategy.LKYH.C[ctype]
		}
	} else {
		t.LHDetail.IsSuppress = true
	}
}

// 高潮涌现结算
func (t *Desk) GCYXSettlement(userid string) {
	// 记录当局打码量
	role := t.roles[userid]
	if role == nil {
		return
	}

	role.LHDStrategy.GCYX.R++
	score := t.ScoreMap[userid].GetSum()
	zlog.Infof("lhd gcyx %s settle: %d", userid, score)
	if score > 0 {
		role.LHDStrategy.GCYX.Win += score
	} else {
		role.LHDStrategy.GCYX.Lose += score
	}
}
