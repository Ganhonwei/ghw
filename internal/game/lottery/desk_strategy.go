package lottery

import (
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sort"
)

func (t *Desk) selectStrategy() {
	// 来玩就赢
	t.triggerLWJY()
	// 来易去难
	t.triggerLYQN()
	//检测策略
	t.checkStrategy()
}

func (t *Desk) checkStrategy() {
	if len(t.CPStrategys) <= 1 {
		return
	}

	sort.Slice(t.CPStrategys, func(i, j int) bool {
		return t.CPStrategys[i].Weight < t.CPStrategys[j].Weight
	})

	strategy := config.GetCPStrategy()

	newStrategy := make([]data.StrategyInfo, 0)
	mutexStrategy := make(map[int]string)

	for _, s := range t.CPStrategys {
		if _, ok := mutexStrategy[s.Id]; ok {
			// 互斥了
			continue
		}
		switch s.Id {
		case LWJY:
			for _, v := range strategy.LWJY.Mutex {
				mutexStrategy[v] = ""
			}
		case LYQN:
			for _, v := range strategy.LYQN.Mutex {
				mutexStrategy[v] = ""
			}
		}
		newStrategy = append(newStrategy, s)
	}

	t.CPStrategys = newStrategy
}

func (t *Desk) triggerLYQN() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetCPStrategy()
	if strategy.Id == 0 {
		glog.Error("no found LYQN strategy config, gameid:", t.GameId)
		return
	}

	if strategy.LYQN.BaseStrategy == nil || strategy.LYQN.O != 1 {
		return
	}

	role := t.roles[userid]

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.LYQN.X
	}

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerCPStrategy(role.User, strategy.LYQN.BaseStrategy, chargeType) ||
		role.Diamond >= strategy.LYQN.D[chargeType] ||
		float64(role.Diamond+int64(role.CashOut))/float64(money) > strategy.LYQN.M[chargeType] {
		return
	}

	T := strategy.LYQN.T[chargeType]

	if T > 0 && (role.CPStrategy.LYQN.TriggerTimes >= T || !utils.RandWan(strategy.LYQN.PS[chargeType])) {
		return
	}

	glog.Info("use LYQN strategy, gameid:", t.GameId)
	t.CPStrategys = append(t.CPStrategys, data.StrategyInfo{Id: LYQN, Weight: strategy.LYQN.Weight})
}

func (t *Desk) triggerLWJY() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetCPStrategy()
	if strategy.Id == 0 {
		glog.Error("no found LWJY strategy config, gameid:", t.GameId)
		return
	}

	if strategy.LWJY.BaseStrategy == nil || strategy.LWJY.O != 1 {
		return
	}

	role := t.roles[userid]

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerCPStrategy(role.User, strategy.LWJY.BaseStrategy, chargeType) ||
		role.Diamond >= strategy.LWJY.M[chargeType] ||
		role.CPStrategy.LWJY.U >= strategy.LWJY.U[chargeType] ||
		role.CPStrategy.LWJY.UZ >= strategy.LWJY.UZ[chargeType] {
		return
	}

	glog.Info("use LWJY strategy, gameid:", t.GameId)
	t.CPStrategys = append(t.CPStrategys, data.StrategyInfo{Id: LWJY, Weight: strategy.LWJY.Weight})
}

func (t *Desk) LYQNStrategy() bool {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetCPStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.LWJY.X
	}

	chargeType := handler.GetChargeType(role.User)

	canOpen := make([]uint32, 0)
	for k, betmap := range t.LSeatRoleBets {
		odds := float64(t.DeskFree.Multiple[k]) / 100
		if b, ok := betmap[userid]; ok {
			w := int64(float64(b.GetSum())*odds) - bet
			if w > 0 {
				// 玩家赢分
				if float64(w+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.LWJY.RR[chargeType] {
					// 超过风控返奖率，不能开
					continue
				}

				if w+bet+role.Diamond+int64(role.CashOut)-money >= strategy.LWJY.PR[chargeType] {
					// 超过风控赢钱上限
					continue
				}
				// 玩家赢分
				canOpen = append(canOpen, k)
			}
		}
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false
	}

	choices := make([]utils.Choice, 0)
	for _, seat := range canOpen {
		choices = append(choices, utils.Choice{Weight: int(SeatProMap[seat]), Item: seat})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false
	}

	t.dispensePoker(choice.Item.(uint32))
	// 最终位置
	return true
}

func (t *Desk) LWJYStrategy() (bool, bool) {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetCPStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.LWJY.X
	}

	chargeType := handler.GetChargeType(role.User)

	if role.CPStrategy.LWJY.WinScore >= strategy.LWJY.C[chargeType] {
		// 超过冷却线，随机开
		// t.StartLead(nil)
		return true, true
	}

	canOpen := make([]uint32, 0)
	for k, betmap := range t.LSeatRoleBets {
		odds := float64(t.DeskFree.Multiple[k]) / 100
		if b, ok := betmap[userid]; ok {
			w := int64(float64(b.GetSum())*odds) - bet
			if w > 0 {
				// 玩家赢分
				if float64(w+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.LWJY.RR[chargeType] {
					// 超过风控返奖率，不能开
					continue
				}

				if w+bet+role.Diamond+int64(role.CashOut)-money >= strategy.LWJY.PR[chargeType] {
					// 超过风控赢钱上限
					continue
				}
				// 玩家赢分
				canOpen = append(canOpen, k)
			}
		}
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, true
	}

	choices := make([]utils.Choice, 0)
	for _, seat := range canOpen {
		choices = append(choices, utils.Choice{Weight: int(SeatProMap[seat]), Item: seat})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, true
	}

	t.dispensePoker(choice.Item.(uint32))
	return true, false
}

func (t *Desk) strategySettlement(userid string) {
	score := t.LScoreMap[userid]

	switch t.CPTriggerStrategy {
	case LWJY:
		t.LWJYSettlement(userid)
	case LYQN:
		t.LYQNSettlement(userid)
	}
	t.eventPost(userid, event.CP_Strategy, &event.CPStrategySettlementEvent{Id: t.CPTriggerStrategy, WinScore: score.GetSum()})
}

func (t *Desk) LYQNSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	role.CPStrategy.LYQN.TriggerTimes++
}

func (t *Desk) LWJYSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	winScore := t.LScoreMap[userid]

	strategy := config.GetCPStrategy()
	ctype := handler.GetChargeType(role.User)

	role.CPStrategy.LWJY.WinScore += winScore.GetSum()

	m := role.Diamond
	if winScore.GetSum() > 0 {
		bet := t.Bets[userid]
		m = m + bet + winScore.GetSum()
	}

	if m >= strategy.LWJY.M[ctype] {
		role.CPStrategy.LWJY.U++
		role.CPStrategy.LWJY.UZ++
		role.CPStrategy.LWJY.WinScore = 0
	}
}
