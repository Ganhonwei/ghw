package plane

import (
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"sort"
)

func (t *Desk) selectStrategy() {
	// 扶摇直上
	t.triggerfyzs()
	// 欲薅无门
	t.triggeryhwm()
	// 虚假情报
	t.triggerxjqb()
	// 起死回生
	t.triggerqshs()
	// 奖池风控
	t.triggerjcfk()
	// 冒险奖励
	t.triggermxjl()
	// 人狂有祸
	t.triggerrkyh()
	// 高潮涌现
	t.triggergcyx()
	// 这个在最后面
	t.checkStrategy()
}

func (t *Desk) checkStrategy() {
	if len(t.CRASHStrategys) <= 1 {
		return
	}

	sort.Slice(t.CRASHStrategys, func(i, j int) bool {
		return t.CRASHStrategys[i].Weight < t.CRASHStrategys[j].Weight
	})

	strategy := config.GetAviatorStrategy()

	newStrategy := make([]data.StrategyInfo, 0)
	mutexStrategy := make(map[int]string)

	for _, s := range t.CRASHStrategys {
		if _, ok := mutexStrategy[s.Id]; ok {
			// 互斥了
			continue
		}
		switch s.Id {
		case FYZS:
			for _, v := range strategy.FYZS.Mutex {
				mutexStrategy[v] = ""
			}
		case YHWM:
			for _, v := range strategy.YHWM.Mutex {
				mutexStrategy[v] = ""
			}
		case QSHS:
			for _, v := range strategy.QSHS.Mutex {
				mutexStrategy[v] = ""
			}
		case JCFK:
			for _, v := range strategy.JCFK.Mutex {
				mutexStrategy[v] = ""
			}
		case MXJL:
			for _, v := range strategy.MXJL.Mutex {
				mutexStrategy[v] = ""
			}
		case RKYH:
			for _, v := range strategy.RKYH.Mutex {
				mutexStrategy[v] = ""
			}
		case GCYX:
			for _, v := range strategy.GCYX.Mutex {
				mutexStrategy[v] = ""
			}
		}
		newStrategy = append(newStrategy, s)
	}

	t.CRASHStrategys = newStrategy
}

// 扶摇直上策略
func (t *Desk) triggerfyzs() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	back := t.CRASHBack[userid]
	bet := t.CrashBets[userid]

	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet.GetSum() <= 0 && back <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found fyzs strategy config, gameid:", t.GameId)
		return
	}

	if strategy.FYZS.BaseStrategy == nil || strategy.FYZS.O != 1 {
		return
	}

	// 玩家类型
	chargeType := handler.GetChargeType(role.User)

	if role, ok := t.roles[userid]; ok {
		if !handler.TiggerCrashStrategy(role.User, strategy.FYZS.BaseStrategy) || // 玩家类型是否满足
			role.AvStrategy.FYZS.TiggerTimes >= strategy.FYZS.U[chargeType] || // 触发次数
			role.Diamond+bet.GetSum() >= strategy.FYZS.M[chargeType] || // 携带上限
			role.AvStrategy.FYZS.AllTiggerTimes >= strategy.FYZS.UZ[chargeType] { // 总触发次数
			// 充过钱了,或者触发过了
			return
		}
	}

	glog.Info("use fyzs strategy, gameid:", t.GameId)
	t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: FYZS, Weight: strategy.FYZS.Weight})
}

// 欲薅无门策略
func (t *Desk) triggeryhwm() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	back := t.CRASHBack[userid]
	bet := t.CrashBets[userid]

	role := t.roles[userid]
	if role == nil {
		return
	}

	if bet.GetSum() <= 0 && back <= 0 {
		// 没下注,不触发
		return
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found yhwm strategy config, gameid:", t.GameId)
		return
	}

	if strategy.YHWM.BaseStrategy == nil || strategy.YHWM.O != 1 {
		return
	}

	if role, ok := t.roles[userid]; ok && role.AvStrategy.YHWM.Tigger {
		// 策略生效中
		glog.Info("use yhwm strategy, gameid:", t.GameId)
		t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: YHWM, Weight: strategy.YHWM.Weight})
	}
}

// 虚假情报策略
func (t *Desk) triggerxjqb() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found xjqb strategy config, gameid:", t.GameId)
		return
	}

	if strategy.XJQB.BaseStrategy == nil || strategy.XJQB.O != 1 {
		return
	}

	role := t.roles[userid]
	if role == nil {
		return
	}

	if handler.TiggerCrashStrategy(role.User, strategy.XJQB.BaseStrategy) {
		// 策略生效中
		glog.Infof("use xjqb strategy, gameid:%s, userid:%s", t.GameId, userid)
		t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: XJQB, Weight: strategy.XJQB.Weight})
	}
}

// 起死回生策略
func (t *Desk) triggerqshs() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found qshs strategy config, gameid:", t.GameId)
		return
	}

	if strategy.QSHS.BaseStrategy == nil || strategy.QSHS.BaseStrategy.O != 1 {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	role := t.roles[userid]
	if role == nil || role.Money <= 0 {
		return
	}

	ctype := handler.GetChargeType(role.User)

	// 援助返奖点
	m := float64(role.CashOut+int32(role.Diamond)) / float64(role.Money)
	if m > strategy.QSHS.M[ctype] {
		return
	}

	if role.Diamond > strategy.QSHS.D[ctype] {
		return
	}

	var maxMultiple, allMultiple, avgMultiple int32
	rounds := role.AvStrategy.QSHS.WinRounds
	if len(role.AvStrategy.QSHS.WinRounds) > strategy.QSHS.R[ctype] {
		rounds = role.AvStrategy.QSHS.WinRounds[len(role.AvStrategy.QSHS.WinRounds)-strategy.QSHS.R[ctype]:]
	}

	for _, w := range rounds {
		if w > maxMultiple {
			maxMultiple = w
		}
		allMultiple += w
	}
	if allMultiple > 0 {
		avgMultiple = allMultiple / int32(len(role.AvStrategy.QSHS.WinRounds))
	}

	n := (float64(avgMultiple)*float64(bet)/100 + float64(role.CashOut) + float64(role.Diamond)) / float64(role.Money)
	a := (float64(maxMultiple)*float64(bet)/100 + float64(role.CashOut) + float64(role.Diamond)) / float64(role.Money)
	if n > strategy.QSHS.N[ctype] || a > strategy.QSHS.A[ctype] {
		return
	}

	if role.AvStrategy.QSHS.TriggerTimes >= strategy.QSHS.T[ctype] && !utils.RandWan(strategy.QSHS.P[ctype]) {
		return
	}

	// 策略生效中
	glog.Infof("use qshs strategy, gameid:%s, userid:%s", t.GameId, userid)
	t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: QSHS, Weight: strategy.QSHS.Weight})
}

// 奖池风控策略
func (t *Desk) triggerjcfk() {
	// userid := t.GetOnlyOneBetPlayer()
	// if userid == "" {
	// 	return
	// }

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found jcfk strategy config, gameid:", t.GameId)
		return
	}

	if strategy.JCFK.BaseStrategy == nil || strategy.JCFK.BaseStrategy.O != 1 {
		return
	}

	trigger := false
	for _, role := range t.roles {
		if role.Robot {
			continue
		}

		ctype := handler.GetChargeType(role.User)

		if bet, ok := t.Bets[role.Userid]; ok && bet > 0 {
			if role.AvStrategy.JCFK.TriggerTimes < strategy.JCFK.T[ctype] {
				return
			}

			money := float64(role.Money)
			if money <= 0 {
				money = float64(strategy.JCFK.N)
			}
			// 总返奖率
			backRate := float64(role.Diamond+bet+int64(role.CashOut)) / money
			if backRate/t.Game.PLANE.BackRate < strategy.JCFK.M[ctype] {
				return
			}
			trigger = true
		}
	}

	// 策略生效中
	if trigger {
		glog.Infof("use jcfk strategy, gameid:%s", t.GameId)
		t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: JCFK, Weight: strategy.JCFK.Weight})
	}
}

// 冒险奖励策略
func (t *Desk) triggermxjl() {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found jcfk strategy config, gameid:", t.GameId)
		return
	}

	if strategy.MXJL.BaseStrategy == nil || strategy.MXJL.BaseStrategy.O != 1 {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	role := t.roles[userid]
	ctype := handler.GetChargeType(role.User)

	money := role.Money
	if money <= 0 {
		money = uint32(strategy.MXJL.X)
	}

	if !handler.TiggerCrashStrategy(role.User, strategy.MXJL.BaseStrategy) ||
		float64(int64(role.CashOut)+role.Diamond+bet)/float64(money) > strategy.MXJL.F[ctype] ||
		len(role.AvStrategy.QSHS.WinRounds) < strategy.MXJL.W[ctype] ||
		len(role.AvStrategy.MXJL.Bets) < strategy.MXJL.R[ctype] ||
		bet < strategy.MXJL.N[ctype] {
		return
	}

	var maxBet int64
	for i := len(role.AvStrategy.MXJL.Bets) - 1; i >= len(role.AvStrategy.MXJL.Bets)-strategy.MXJL.R[ctype]; i-- {
		maxBet += role.AvStrategy.MXJL.Bets[i]
	}
	if maxBet/int64(strategy.MXJL.R[ctype]) < strategy.MXJL.C[ctype] {
		return
	}

	var maxMultiple int32
	for i := len(role.AvStrategy.QSHS.WinRounds) - 1; i >= len(role.AvStrategy.QSHS.WinRounds)-strategy.MXJL.W[ctype]; i-- {
		maxMultiple += role.AvStrategy.QSHS.WinRounds[i]
	}
	avgMul := float64(maxMultiple/int32(strategy.MXJL.W[ctype])) / 100
	if avgMul < strategy.MXJL.M[ctype] {
		return
	}

	if role.AvStrategy.MXJL.TriggerTimes >= strategy.MXJL.T[ctype] && !utils.RandWan(int32(strategy.MXJL.P[ctype])) {
		return
	}

	// 策略生效中
	glog.Infof("use mxjl strategy, gameid:%s", t.GameId)
	t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: MXJL, Weight: strategy.MXJL.Weight})
}

// 人狂有祸
func (t *Desk) triggerrkyh() {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found jcfk strategy config, gameid:", t.GameId)
		return
	}

	if strategy.RKYH.BaseStrategy == nil || strategy.RKYH.BaseStrategy.O != 1 {
		return
	}

	var userid string
	trigger := false
	for _, role := range t.roles {
		if role.Robot {
			continue
		}
		ctype := handler.GetChargeType(role.User)
		userid = role.Userid
		money := role.Money
		if money == 0 {
			// 未付费玩家默充金额x
			money = uint32(strategy.RKYH.X)
		}
		pu := (float64(role.CashOut) + float64(role.Diamond)) / float64(money)
		if pu >= strategy.RKYH.PU[ctype] && handler.TiggerCrashStrategy(role.User, strategy.RKYH.BaseStrategy) {
			trigger = true
		}
	}

	// 策略生效中
	if trigger {
		glog.Infof("use rkyh strategy, gameid:%s, %s", t.GameId, userid)
		t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: RKYH, Weight: strategy.RKYH.Weight})
	}
}

// 高潮涌现
func (t *Desk) triggergcyx() {

	//记录打码量
	for _, role := range t.roles {
		if role.Robot {
			continue
		}
		userid := role.Userid

		if bet, ok := t.Bets[role.Userid]; ok && bet > 0 {
			role.AvStrategy.GCYX.DMRecord = append(role.AvStrategy.GCYX.DMRecord, bet)
			if len(role.AvStrategy.GCYX.DMRecord) > 50 { //最近50局
				role.AvStrategy.GCYX.DMRecord = role.AvStrategy.GCYX.DMRecord[1:]
			}
			t.eventPost(userid, event.Aviator_GCYX_BET, &event.AviatorGCYXBetEvent{Bet: bet})
		}
	}

	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		glog.Error("no found gcyx strategy config, gameid:", t.GameId)
		return
	}

	if strategy.GCYX.BaseStrategy == nil || strategy.GCYX.BaseStrategy.O != 1 {
		return
	}

	var userid string
	trigger := false

	for _, role := range t.roles {
		if role.Robot {
			continue
		}
		ctype := handler.GetChargeType(role.User)
		userid = role.Userid
		_ = ctype

		if !handler.TiggerCrashStrategy(role.User, strategy.GCYX.BaseStrategy) {
			return
		}

		if bet, ok := t.Bets[role.Userid]; ok && bet > 0 {

			needJudge := true

			//贤者状态，不检测
			if role.AvStrategy.GCYX.IsStateXZ {
				return
			}

			//如果处于高潮状态，不需要再判断是否触发,直接触发
			if role.AvStrategy.GCYX.IsStateGC {
				needJudge = false
				trigger = true
			}

			if needJudge { //不处于高潮状态，判断是否触发，再触发
				//单日高潮次数上限
				if role.AvStrategy.GCYX.DailyGCTimes >= int(strategy.GCYX.S[ctype]) {
					return
				}

				money := float64(role.Money)
				if money <= 0 {
					money = float64(strategy.GCYX.X)
				}
				//条件1：
				// 计算总返奖率
				backRate := (float64(role.Diamond) + float64(role.CashOut)) / money
				// 检查总返奖率是否低于禁止高潮返奖率
				if backRate >= strategy.GCYX.FP[ctype] {
					return
				}

				//条件2 计算概率中不中
				num := strategy.GCYX.HP[ctype] + int32(role.AvStrategy.GCYX.EvoTimes)*strategy.GCYX.HP1[ctype]
				if !utils.RandWan(num) {
					return
				}

				//计算均码量
				f := strategy.GCYX.F[ctype]             //前戏局数
				l := len(role.AvStrategy.GCYX.DMRecord) //记录条数
				average := int64(0)

				if l <= int(f) { //记录条数不足或刚好等于
					sum := int64(0)
					for _, v := range role.AvStrategy.GCYX.DMRecord {
						sum += v
					}
					average = sum / int64(l)
				} else { //记录条数大于所需条数
					count := 0
					sum := int64(0)
					for i := l - 1; i >= 0; i-- {
						sum += role.AvStrategy.GCYX.DMRecord[i]
						count++
						if count >= int(f) {
							break
						}
					}
					average = sum / int64(f)
				}

				role.AvStrategy.GCYX.IsStateGC = true //进入高潮模式
				role.AvStrategy.GCYX.M++              //增加高潮次数m
				role.AvStrategy.GCYX.DailyGCTimes++   //增加每日高潮次数
				role.AvStrategy.GCYX.AvgBet = average //更新均码量
				t.eventPost(role.Userid, event.Aviator_GCYX_Trigger, &event.AviatorGCYXTriggerEvent{AvgBet: average})

				trigger = true
			}
		}
	}

	if trigger {
		glog.Infof("use gcyx strategy, gameid:%s, %s", t.GameId, userid)
		t.CRASHStrategys = append(t.CRASHStrategys, data.StrategyInfo{Id: GCYX, Weight: -1})
	}

}

func (t *Desk) fyzsStrategy() bool {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return true
	}

	role := t.roles[userid]
	strategy := config.GetAviatorStrategy()
	ctype := handler.GetChargeType(role.User)

	// 是否在策略冷却中
	_, back := t.CRASHBack[userid]
	if role.AvStrategy.FYZS.WinScore >= strategy.FYZS.C[ctype] && !back {
		// 不干预
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}

	bet := t.Bets[userid]
	money := int64(role.Money)
	if money <= 0 {
		money = strategy.FYZS.X
	}

	// 必爆点G
	g1 := (strategy.FYZS.RR[ctype]*float64(money) - float64(role.CashOut) - float64(role.Diamond)) / float64(bet)
	g2 := float64(strategy.FYZS.PR[ctype]+money-int64(role.CashOut)-role.Diamond) / float64(bet)
	g := int32(math.Min(g1, g2) * 100)
	if g1 > 1 && g2 < 1 {
		g = int32(g1 * 100)
	}
	if g1 < 1 && g2 > 1 {
		g = int32(g2 * 100)
	}

	if t.CRASHMulpitle >= g {
		return true
	}

	// 没撤离，不爆
	if back, ok := t.CRASHBack[userid]; !ok {
		if t.CRASHMulpitle <= int32(strategy.FYZS.H[ctype]*100) {
			return false
		} else {
			// 修正值
			q := strategy.FYZS.Q * strategy.FYZS.H[ctype]
			return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
		}
	} else {
		// 已经撤离了,检测爆炸
		// 修正值
		q := strategy.FYZS.Q * float64(back) / 100
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
	}
}

func (t *Desk) yhwmStrategy() bool {
	strategy := config.GetAviatorStrategy()
	userid := t.GetOnlyOneBetPlayer()

	role := t.roles[userid]
	if role == nil {
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}

	ctype := handler.GetChargeType(role.User)

	// 1.0瞬爆
	if t.CRASHMulpitle == 100 {
		if utils.RandWan(strategy.YHWM.P[ctype]) {
			return true
		} else {
			return false
		}
	}

	return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
}

func (t *Desk) xjqbStrategy() (bool, bool) {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return false, false
	}

	role := t.roles[userid]
	if role == nil {
		return false, false
	}

	strategy := config.GetAviatorStrategy()

	ctype := handler.GetChargeType(role.User)

	// 是否观察局
	if _, ok := t.Bets[userid]; !ok {
		return true, handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, strategy.XJQB.P1[ctype], 0, r)
	} else if _, ok := t.CRASHBack[userid]; ok {
		return true, handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, strategy.XJQB.P2[ctype], 0, r)
	}

	return false, false
}

func (t *Desk) qshsStrategy() bool {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}

	strategy := config.GetAviatorStrategy()

	role := t.roles[userid]

	bet := t.Bets[userid]
	ctype := handler.GetChargeType(role.User)

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.QSHS.X
	}

	// 计算必爆点
	g1 := (strategy.QSHS.RR[ctype]*float64(money) - float64(role.CashOut) - float64(role.Diamond)) / float64(bet)
	g2 := float64(strategy.QSHS.PF[ctype]+money-int64(role.CashOut)-role.Diamond) / float64(bet)
	g := int32(math.Min(g1, g2) * 100)
	if g1 > 1 && g2 < 1 {
		g = int32(g1 * 100)
	}
	if g1 < 1 && g2 > 1 {
		g = int32(g2 * 100)
	}

	if t.CRASHMulpitle >= g {
		return true
	}

	// 没撤离，不爆
	if back, ok := t.CRASHBack[userid]; !ok {
		if t.CRASHMulpitle <= int32(strategy.QSHS.S[ctype]*100) {
			return false
		} else {
			// 修正值
			q := strategy.QSHS.Q * strategy.QSHS.S[ctype]
			return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
		}
	} else {
		// 已经撤离了,检测爆炸
		// 修正值
		q := strategy.QSHS.Q * float64(back) / 100
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
	}
}

func (t *Desk) jcfkStrategy() (bool, bool) {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return false, handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}

	strategy := config.GetAviatorStrategy()

	role := t.roles[userid]

	bet := t.Bets[userid]

	ctype := handler.GetChargeType(role.User)

	if t.CRASHMulpitle < int32(strategy.JCFK.B[ctype]*100) {
		return false, false
	}

	// 总返奖率
	var backRate float64
	if role.Money > 0 {
		backRate = float64(role.Diamond+bet+int64(role.CashOut)) / float64(role.Money)
	} else {
		backRate = float64(role.Diamond+bet+int64(role.CashOut)) / float64(strategy.JCFK.N)
	}

	// 修正值
	q := strategy.JCFK.Q * (backRate / t.Game.PLANE.BackRate)
	return true, handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
}

// 冒险奖励
func (t *Desk) mxjlStrategy() bool {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}

	strategy := config.GetAviatorStrategy()

	role := t.roles[userid]
	ctype := handler.GetChargeType(role.User)

	bet := t.Bets[userid]

	var money float64 = float64(strategy.MXJL.X)
	if role.Money > 0 {
		money = float64(role.Money)
	}

	// 计算必爆点
	g1 := (strategy.MXJL.RR[ctype]*money - float64(role.CashOut) - float64(role.Diamond)) / float64(bet)
	g2 := float64(strategy.MXJL.PF[ctype]+int64(money)-int64(role.CashOut)-role.Diamond) / float64(bet)
	g := int32(math.Min(g1, g2) * 100)
	if g1 > 1 && g2 < 1 {
		g = int32(g1 * 100)
	}
	if g1 < 1 && g2 > 1 {
		g = int32(g2 * 100)
	}

	if t.CRASHMulpitle >= g {
		return true
	}

	// 没撤离，不爆
	if back, ok := t.CRASHBack[userid]; !ok {
		if t.CRASHMulpitle <= int32(strategy.MXJL.S[ctype]*100) {
			return false
		} else {
			// 修正值
			q := strategy.MXJL.Q * strategy.MXJL.S[ctype]
			return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
		}
	} else {
		// 已经撤离了,检测爆炸
		// 修正值
		q := strategy.MXJL.Q * float64(back) / 100
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, q, r)
	}
}

// 人狂有祸
func (t *Desk) rkyhStrategy() (bool, bool) {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return false, handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, t.Game.PLANE.BackRate, 0, r)
	}
	player := t.getPlayer(userid)
	ctype := handler.GetChargeType(player)

	strategy := config.GetAviatorStrategy()

	// 瞬爆概率p1在1倍爆炸
	if t.DeskFree.CRASHMulpitle == 100 {
		if utils.RandWan(strategy.RKYH.P1[ctype]) {
			return true, true
		} else {
			return true, false
		}
	}

	if handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, strategy.RKYH.P7[ctype], 0, r) {
		return true, true
	}

	// m=算前w局的平均值
	var winMSum, count int32
	length := len(player.AvStrategy.RKYH.WinMulpitles)
	for i := length - 1; i >= 0; i-- {
		if count >= int32(strategy.RKYH.W[ctype]) {
			break
		}
		winMSum += player.AvStrategy.RKYH.WinMulpitles[i]
		count++
	}
	// w局局均逃跑倍数m
	var m int32
	if count > 0 {
		m = winMSum / count
	}
	bet := t.Bets[userid]
	if m > 0 && bet >= strategy.RKYH.C[ctype] {
		// 有天罚概率p2在m-n倍的时候直接爆炸
		if t.CRASHMulpitle >= int32(m) && t.CRASHMulpitle <= int32(strategy.RKYH.N[ctype]*100) {
			if utils.RandWan(strategy.RKYH.P2[ctype]) {
				// 记录r肥割
				t.CRASHDetail.CRASHDetail.RkyhRfge = true
				return true, true
			} else {
				return true, false
			}
		}
	}

	return true, false
}

// 高潮涌现
func (t *Desk) gcyxStrategy() bool {
	userid := t.GetOnlyOneBetPlayer()
	if userid == "" {
		return true
	}

	role := t.roles[userid]
	strategy := config.GetAviatorStrategy()
	ctype := handler.GetChargeType(role.User)

	//没撤离，不爆
	if _, ok := t.CRASHBack[userid]; !ok {
		multiple := t.CRASHMulpitle
		//没飞过不爆倍数上限h
		if multiple <= int32(strategy.GCYX.H[ctype]*100) {
			return false
		} else {
			return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, strategy.GCYX.P0[ctype], 0, r)
		}
	} else {
		// 已经撤离了,检测爆炸
		return handler.CrashIsBoom(t.DeskFree.CRASHMulpitle, strategy.GCYX.P0[ctype], 0, r)
	}
}

func (t *Desk) yhwmNoTrigger(userid string) {
	score := t.CRASHScoreMap[userid]
	strategy := config.GetAviatorStrategy()

	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		ctype := handler.GetChargeType(role.User)

		if !role.AvStrategy.YHWM.Tigger && (role.AvStrategy.YHWM.TiggerTimes < strategy.YHWM.T[ctype] || strategy.YHWM.T[ctype] == -1) {
			if score.GetSum() > 0 {
				// 策略生效前赢分
				role.AvStrategy.YHWM.MonitorRounds++
				role.AvStrategy.YHWM.AviatorMulpitles = append(role.AvStrategy.YHWM.AviatorMulpitles, t.CRASHBack[userid])
				role.AvStrategy.YHWM.WinScore += score.GetSum()
			} else {
				role.AvStrategy.YHWM = data.AviatorYHWM{
					TiggerTimes:    role.AvStrategy.YHWM.TiggerTimes,
					AllRecyleScore: role.AvStrategy.YHWM.AllRecyleScore,
				}
			}

			yhwm := role.AvStrategy.YHWM
			if (yhwm.TiggerTimes < strategy.YHWM.T[ctype] || strategy.YHWM.T[ctype] == -1) &&
				yhwm.MonitorRounds >= strategy.YHWM.R[ctype] && len(yhwm.AviatorMulpitles) > 0 && // 监控局数满足条件
				handler.TiggerCrashStrategy(role.User, strategy.YHWM.BaseStrategy) { // 玩家类型是否满足
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
				if middle <= strategy.YHWM.M[ctype] && (avg <= strategy.YHWM.A[ctype] || strategy.YHWM.A[ctype] == -1) {
					role.AvStrategy.YHWM.Tigger = true
					role.AvStrategy.YHWM.TiggerTimes++
					// role.AvStrategy.YHWM.AllRecyleScore += role.AvStrategy.YHWM.RecyleScore
					role.AvStrategy.YHWM.RecyleScore = 0
				}
			}
			// 事件
			// t.eventPost(role.Userid, event.Crash_Settlement, &event.CrashSettlementEvent{Score: score.GetSum(), Multiple: t.CRASHBack[userid]})
		}
	}
}

func (t *Desk) qshsTrigger(userid string) {
	// 起死回生
	if s, ok := t.CRASHScoreMap[userid]; ok {
		if user, ok := t.roles[userid]; ok {
			if s.GetSum() > 0 {
				user.AvStrategy.QSHS.WinRounds = append(user.AvStrategy.QSHS.WinRounds, t.CRASHMulpitle)
				if len(user.AvStrategy.QSHS.WinRounds) > 100 {
					user.AvStrategy.QSHS.WinRounds = user.AvStrategy.QSHS.WinRounds[len(user.AvStrategy.QSHS.WinRounds)-100:]
				}
			}
		}
	}
}

func (t *Desk) jcfkTrigger(userid string) {
	if r, ok := t.roles[userid]; ok {
		// 奖池风控
		if t.CRASHMulpitle > 2000 {
			r.AvStrategy.JCFK.TriggerTimes++
		}
	}
}

func (t *Desk) mxjlTrigger(userid string) {
	// 冒险奖励
	if s, ok := t.Bets[userid]; ok {
		if user, ok := t.roles[userid]; ok {
			user.AvStrategy.MXJL.Bets = append(user.AvStrategy.MXJL.Bets, s)
			if len(user.AvStrategy.MXJL.Bets) > 100 {
				user.AvStrategy.MXJL.Bets = user.AvStrategy.MXJL.Bets[len(user.AvStrategy.MXJL.Bets)-100:]
			}
		}
	}
}

func (t *Desk) gcyxTrigger(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}

	// 高潮涌现
	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		ctype := handler.GetChargeType(role.User)

		//贤者状态，不计数
		if role.AvStrategy.GCYX.IsStateXZ {
			role.AvStrategy.GCYX.XZTimes++
			if role.AvStrategy.GCYX.XZTimes >= int(strategy.GCYX.C[ctype]) {
				role.AvStrategy.GCYX.IsStateXZ = false
				role.AvStrategy.GCYX.XZTimes = 0
				t.eventPost(userid, event.Aviator_GCXY_XZ_Over, &event.AviatorGCXYXZOverEvent{IsOver: true})
			} else {
				t.eventPost(userid, event.Aviator_GCXY_XZ_Over, &event.AviatorGCXYXZOverEvent{IsOver: false})
			}
			return
		}

		role.AvStrategy.GCYX.EvoTimes++
	}
}

// 策略结算
func (t *Desk) strategySettlement(userid string) {
	// 欲薅无门没生效的情况
	t.yhwmNoTrigger(userid)
	// 起死回生
	t.qshsTrigger(userid)
	// 奖池风控
	t.jcfkTrigger(userid)
	// 冒险奖励
	t.mxjlTrigger(userid)
	// 人狂有祸
	t.rkyhTrigger(userid)
	// 高潮涌现
	t.gcyxTrigger(userid)

	for _, s := range t.CRASHTriggerStrategys {
		switch s.Id {
		case FYZS:
			t.fyzsSettlement(userid)
		case YHWM:
			t.yhwmSettlement(userid)
		case QSHS:
			t.qshsSettlement(userid)
		case MXJL:
			t.mxjlSettlement(userid)
		case GCYX:
			t.gcyxSettlement(userid)
		}
	}
}

// 高潮涌现策略结算
func (t *Desk) gcyxSettlement(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}
	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		ctype := handler.GetChargeType(role.User)
		score := t.CRASHScoreMap[userid]

		role.AvStrategy.GCYX.WinScore += score.GetSum()

		t.eventPost(userid, event.Aviator_GCYX_SCORE, &event.AviatorGCYXScoreEvent{Score: score.GetSum()})

		//判断是否结束高潮状态
		g := strategy.GCYX.G[ctype] //高潮盈利倍率上限

		if role.AvStrategy.GCYX.WinScore >= int64(g)*role.AvStrategy.GCYX.AvgBet {
			role.AvStrategy.GCYX.IsStateGC = false //退出高潮状态
			role.AvStrategy.GCYX.IsStateXZ = true  //进入贤者状态
			role.AvStrategy.GCYX.EvoTimes = 0      //重置n
			role.AvStrategy.GCYX.WinScore = 0      //重置赢分
			role.AvStrategy.GCYX.AvgBet = 0        //重置均码量

			t.eventPost(userid, event.Aviator_GCXY_GC_Over, &event.AviatorGCXYGCOverEvent{})
		}
	}
}

// 扶摇直上策略结算
func (t *Desk) fyzsSettlement(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}

	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}
		role.AvStrategy.FYZS.EvoTimes++
		role.AvStrategy.FYZS.AllEvoTimes++

		// 是否终止策略
		ctype := handler.GetChargeType(role.User)
		if role.Diamond >= strategy.FYZS.M[ctype] {
			role.AvStrategy.FYZS.TiggerTimes++
			// role.AvStrategy.FYZS.AllTiggerTimes++
			// 事件
			// t.eventPost(userid, event.Crash_FYZS_Tigger, &event.CrashFYZSTiggerEvent{Ttype: 1})
		}

		score := t.CRASHScoreMap[userid]
		role.AvStrategy.FYZS.WinScore += score.GetSum()
		// 事件
		t.eventPost(userid, event.Aviator_FYZS, &event.AviatorFYZSEvent{Score: score.GetSum()})
	}
}

// 人狂有祸记录
func (t *Desk) rkyhTrigger(userid string) {
	// bet := t.Bets[userid]
	// 记录赢局记录逃跑倍数
	cur := t.CRASHScoreMap[userid]
	e := new(event.AviatorRKYHTiggerEvent)
	role, ok := t.roles[userid]
	if !ok {
		return
	}

	ctype := handler.GetChargeType(role.User)

	back := t.CRASHBack[userid]
	if cur.GetSum() > 0 && back > 0 {
		role.AvStrategy.RKYH.WinMulpitles = append(role.AvStrategy.RKYH.WinMulpitles, back)

		w := config.GetAviatorStrategy().RKYH.W[ctype]
		if w <= 0 {
			w = 10
		}
		length := len(role.AvStrategy.RKYH.WinMulpitles)
		if length > int(w) {
			role.AvStrategy.RKYH.WinMulpitles = role.AvStrategy.RKYH.WinMulpitles[length-int(w):]
		}
		e.WinMulti = back
		e.W = int32(w)
	}

	for _, s := range t.CRASHTriggerStrategys {
		if s.Id == RKYH {
			e.Tigger = true
			role.AvStrategy.RKYH.TriggerTimes++
			break
		}
	}
	if e.WinMulti > 0 || e.Tigger {
		t.eventPost(userid, event.Aviator_RKYH_Tigger, e)
	}
}

// 欲薅无门策略结算
func (t *Desk) yhwmSettlement(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}

	score := t.CRASHScoreMap[userid]

	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		ctype := handler.GetChargeType(role.User)

		if score.GetSum() <= 0 {
			if role.AvStrategy.YHWM.Tigger {
				// 策略生效后回收分数
				role.AvStrategy.YHWM.RecyleScore -= score.GetSum()
				role.AvStrategy.YHWM.AllRecyleScore -= score.GetSum()
			} else {
				role.AvStrategy.YHWM = data.AviatorYHWM{
					TiggerTimes:    role.AvStrategy.YHWM.TiggerTimes,
					AllRecyleScore: role.AvStrategy.YHWM.AllRecyleScore,
				}
			}
		}

		if role.AvStrategy.YHWM.Tigger &&
			float64(role.AvStrategy.YHWM.RecyleScore) >= float64(role.AvStrategy.YHWM.WinScore)*strategy.YHWM.X[ctype] {
			// 生效中,判断是否要取消
			role.AvStrategy.YHWM = data.AviatorYHWM{
				TiggerTimes:    role.AvStrategy.YHWM.TiggerTimes,
				AllRecyleScore: role.AvStrategy.YHWM.AllRecyleScore,
			}
		}
	}
}

// 起死回生
func (t *Desk) qshsSettlement(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}

	role := t.roles[userid]
	if role == nil {
		return
	}

	role.AvStrategy.QSHS.TriggerTimes++
	// 事件
	t.eventPost(userid, event.Aviator_QSHS, new(event.AviatorQSHSEvent))
}

func (t *Desk) mxjlSettlement(userid string) {
	strategy := config.GetAviatorStrategy()
	if strategy.Id == 0 {
		return
	}

	role := t.roles[userid]
	if role == nil {
		return
	}

	role.AvStrategy.MXJL.TriggerTimes++
	// 事件
	t.eventPost(userid, event.Aviator_MXJL, new(event.AviatorMXJLEvent))
}
