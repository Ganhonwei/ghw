package andarbahar

import (
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"sort"
)

func (t *Desk) selectStrategy() {
	// 安能求死
	t.triggerANQS()
	// 安然躺赢
	t.triggerARTY()
	//检测策略
	t.checkStrategy()
}

func (t *Desk) checkStrategy() {
	if len(t.ABStrategys) <= 1 {
		return
	}

	sort.Slice(t.ABStrategys, func(i, j int) bool {
		return t.ABStrategys[i].Weight < t.ABStrategys[j].Weight
	})

	strategy := config.GetABStrategy()

	newStrategy := make([]data.StrategyInfo, 0)
	mutexStrategy := make(map[int]string)

	for _, s := range t.ABStrategys {
		if _, ok := mutexStrategy[s.Id]; ok {
			// 互斥了
			continue
		}
		switch s.Id {
		case ANQS:
			for _, v := range strategy.ANQS.Mutex {
				mutexStrategy[v] = ""
			}
		case ARTY:
			for _, v := range strategy.ARTY.Mutex {
				mutexStrategy[v] = ""
			}
		}
		newStrategy = append(newStrategy, s)
	}

	t.ABStrategys = newStrategy
}

func (t *Desk) triggerANQS() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetABStrategy()
	if strategy.Id == 0 {
		glog.Error("no found anqs strategy config, gameid:", t.GameId)
		return
	}

	if strategy.ANQS.BaseStrategy == nil || strategy.ANQS.O != 1 {
		return
	}

	role := t.roles[userid]

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.ANQS.X
	}

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerABStrategy(role.User, strategy.ANQS.BaseStrategy, chargeType) ||
		role.Diamond >= strategy.ANQS.D[chargeType] ||
		float64(role.Diamond+int64(role.CashOut))/float64(money) > strategy.ANQS.M[chargeType] {
		return
	}

	T := strategy.ANQS.T[chargeType]

	if T > 0 && (role.ABStrategy.ANQS.TriggerTimes >= T || !utils.RandWan(strategy.ANQS.PS[chargeType])) {
		return
	}

	glog.Info("use ANQS strategy, gameid:", t.GameId)
	t.ABStrategys = append(t.ABStrategys, data.StrategyInfo{Id: ANQS, Weight: strategy.ANQS.Weight})
}

func (t *Desk) triggerARTY() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetABStrategy()
	if strategy.Id == 0 {
		glog.Error("no found ARTY strategy config, gameid:", t.GameId)
		return
	}

	if strategy.ARTY.BaseStrategy == nil || strategy.ARTY.O != 1 {
		return
	}

	role := t.roles[userid]

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerABStrategy(role.User, strategy.ARTY.BaseStrategy, chargeType) ||
		role.Diamond+bet > strategy.ARTY.M[chargeType] ||
		role.ABStrategy.ARTY.U >= strategy.ARTY.U[chargeType] ||
		role.ABStrategy.ARTY.UZ >= strategy.ARTY.UZ[chargeType] {
		return
	}

	glog.Info("use ARTY strategy, gameid:", t.GameId)
	t.ABStrategys = append(t.ABStrategys, data.StrategyInfo{Id: ARTY, Weight: strategy.ARTY.Weight})
}

func (t *Desk) ANQSStrategy() (bool, int) {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetABStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.ANQS.X
	}

	chargeType := handler.GetChargeType(role.User)

	canOpen := make([]DrawResult, 0)

	wins := t.calWinningScoreOf16()
	for _, w := range wins {
		if w.Score <= 0 {
			// 玩家赢分
			if float64(-w.Score+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.ANQS.RR[chargeType] {
				// 超过风控返奖率，不能开
				continue
			}

			if -w.Score+bet+role.Diamond+int64(role.CashOut)-money >= strategy.ANQS.PR[chargeType] {
				// 超过风控赢钱上限
				continue
			}
			canOpen = append(canOpen, w)
		}
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0
	}

	m := math.Pow10(8)
	choices := make([]utils.Choice, 0)
	for _, v := range canOpen {
		mainPro := t.getSeatProbility(v.Winner)
		sidePro := t.getSeatProbility(v.SideWinner)

		choices = append(choices, utils.Choice{Weight: int(mainPro * sidePro * m), Item: v})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0
	}

	// 最终位置
	reslut := choice.Item.(DrawResult)
	return true, algo.ABDispaterTimes(reslut.Winner, reslut.SideWinner)
}

func (t *Desk) ARTYStrategy() (bool, int) {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetABStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.ARTY.X
	}

	chargeType := handler.GetChargeType(role.User)

	if role.ABStrategy.ARTY.WinScore >= strategy.ARTY.C[chargeType] {
		// 超过冷却线，随机开
		return true, 0
	}

	canOpen := make([]DrawResult, 0)

	wins := t.calWinningScoreOf16()
	for _, w := range wins {
		if w.Score > 0 {
			continue
		}
		// 玩家赢分
		if float64(-w.Score+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.ARTY.RR[chargeType] {
			// 超过风控返奖率，不能开
			continue
		}

		if -w.Score+bet+role.Diamond+int64(role.CashOut)-money >= strategy.ARTY.PR[chargeType] {
			// 超过风控赢钱上限
			continue
		}
		canOpen = append(canOpen, w)
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0
	}

	m := math.Pow10(8)
	choices := make([]utils.Choice, 0)
	for _, v := range canOpen {
		mainPro := t.getSeatProbility(v.Winner)
		sidePro := t.getSeatProbility(v.SideWinner)

		choices = append(choices, utils.Choice{Weight: int(mainPro * sidePro * m), Item: v})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0
	}

	// 最终位置
	reslut := choice.Item.(DrawResult)
	return true, algo.ABDispaterTimes(reslut.Winner, reslut.SideWinner)
}

func (t *Desk) strategySettlement(userid string) {
	score := t.ABScoreMap[userid]

	for _, s := range t.ABStrategys {
		switch s.Id {
		case ANQS:
			t.ANQSSettlement(userid)
		case ARTY:
			t.ARTYSettlement(userid)
		}

		t.eventPost(userid, event.AB_Strategy, &event.ABStrategySettlementEvent{Id: s.Id, WinScore: score.GetSum()})
	}
}

func (t *Desk) ANQSSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	role.ABStrategy.ANQS.TriggerTimes++
}

func (t *Desk) ARTYSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	winScore := t.ABScoreMap[userid]

	strategy := config.GetABStrategy()
	ctype := handler.GetChargeType(role.User)

	role.ABStrategy.ARTY.WinScore += winScore.GetSum()

	if role.Diamond >= strategy.ARTY.M[ctype] {
		role.ABStrategy.ARTY.U++
		role.ABStrategy.ARTY.WinScore = 0
	}
}
