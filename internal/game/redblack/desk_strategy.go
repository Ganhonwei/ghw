package redblack

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
	// 红运当头
	t.triggerHYDT()
	// 绝处逢生
	t.triggerJCFS()
	//检测策略
	t.checkStrategy()
}

func (t *Desk) checkStrategy() {
	if len(t.RBStrategys) <= 1 {
		return
	}

	sort.Slice(t.RBStrategys, func(i, j int) bool {
		return t.RBStrategys[i].Weight < t.RBStrategys[j].Weight
	})

	strategy := config.GetRBStrategy()

	newStrategy := make([]data.StrategyInfo, 0)
	mutexStrategy := make(map[int]string)

	for _, s := range t.RBStrategys {
		if _, ok := mutexStrategy[s.Id]; ok {
			// 互斥了
			continue
		}
		switch s.Id {
		case HYDT:
			for _, v := range strategy.HYDT.Mutex {
				mutexStrategy[v] = ""
			}
		case JCFS:
			for _, v := range strategy.JCFS.Mutex {
				mutexStrategy[v] = ""
			}
		}
		newStrategy = append(newStrategy, s)
	}

	t.RBStrategys = newStrategy
}

func (t *Desk) triggerJCFS() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetRBStrategy()
	if strategy.Id == 0 {
		glog.Error("no found JCFS strategy config, gameid:", t.GameId)
		return
	}

	if strategy.JCFS.BaseStrategy == nil || strategy.JCFS.O != 1 {
		return
	}

	role := t.roles[userid]

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.JCFS.X
	}

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerRBStrategy(role.User, strategy.JCFS.BaseStrategy, chargeType) ||
		role.Diamond >= strategy.JCFS.D[chargeType] ||
		float64(role.Diamond+int64(role.CashOut))/float64(money) > strategy.JCFS.M[chargeType] {
		return
	}

	T := strategy.JCFS.T[chargeType]

	if T > 0 && (role.RBStrategy.JCFS.TriggerTimes >= T || !utils.RandWan(strategy.JCFS.PS[chargeType])) {
		return
	}

	glog.Info("use JCFS strategy, gameid:", t.GameId)
	t.RBStrategys = append(t.RBStrategys, data.StrategyInfo{Id: JCFS, Weight: strategy.JCFS.Weight})
}

func (t *Desk) triggerHYDT() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	strategy := config.GetRBStrategy()
	if strategy.Id == 0 {
		glog.Error("no found HYDT strategy config, gameid:", t.GameId)
		return
	}

	if strategy.HYDT.BaseStrategy == nil || strategy.HYDT.O != 1 {
		return
	}

	role := t.roles[userid]

	chargeType := handler.GetChargeType(role.User)

	if !handler.TiggerRBStrategy(role.User, strategy.HYDT.BaseStrategy, chargeType) ||
		role.Diamond+bet > strategy.HYDT.M[chargeType] ||
		role.RBStrategy.HYDT.U >= strategy.HYDT.U[chargeType] ||
		role.RBStrategy.HYDT.UZ >= strategy.HYDT.UZ[chargeType] {
		return
	}

	glog.Info("use HYDT strategy, gameid:", t.GameId)
	t.RBStrategys = append(t.RBStrategys, data.StrategyInfo{Id: HYDT, Weight: strategy.HYDT.Weight})
}

func (t *Desk) JCFSStrategy() (bool, uint32, uint32) {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetRBStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.JCFS.X
	}

	chargeType := handler.GetChargeType(role.User)

	canOpen := make([]DrawResult, 0)

	wins := t.calWinningScoreOf12()
	for _, w := range wins {
		if w.Score <= 0 {
			// 玩家赢分
			if float64(-w.Score+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.JCFS.RR[chargeType] {
				// 超过风控返奖率，不能开
				continue
			}

			if -w.Score+bet+role.Diamond+int64(role.CashOut)-money >= strategy.JCFS.PR[chargeType] {
				// 超过风控赢钱上限
				continue
			}
			canOpen = append(canOpen, w)
		}
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0, 0
	}

	m := math.Pow10(8)
	choices := make([]utils.Choice, 0)
	for _, v := range canOpen {
		probility := t.getSeatProbility(v.Winner, v.WinType)
		choices = append(choices, utils.Choice{Weight: int(probility * m), Item: v})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0, 0
	}

	// 最终位置
	reslut := choice.Item.(DrawResult)
	return true, reslut.Winner, reslut.WinType
}

func (t *Desk) HYDTStrategy() (bool, uint32, uint32) {
	userid := t.GetOnlyOnePlayer()

	role := t.roles[userid]
	bet := t.Bets[userid]

	strategy := config.GetRBStrategy()

	money := int64(role.Money)
	if money <= 0 {
		money = strategy.HYDT.X
	}

	chargeType := handler.GetChargeType(role.User)

	if role.RBStrategy.HYDT.WinScore >= strategy.HYDT.C[chargeType] {
		// 超过冷却线，随机开
		return true, 0, 0
	}

	canOpen := make([]DrawResult, 0)

	wins := t.calWinningScoreOf12()
	for _, w := range wins {
		if w.Score > 0 {
			continue
		}
		// 玩家赢分
		if float64(-w.Score+bet+role.Diamond+int64(role.CashOut))/float64(money) >= strategy.HYDT.RR[chargeType] {
			// 超过风控返奖率，不能开
			continue
		}

		if -w.Score+bet+role.Diamond+int64(role.CashOut)-money >= strategy.HYDT.PR[chargeType] {
			// 超过风控赢钱上限
			continue
		}
		canOpen = append(canOpen, w)
	}

	if len(canOpen) == 0 {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0, 0
	}

	m := math.Pow10(8)
	choices := make([]utils.Choice, 0)
	for _, v := range canOpen {
		probility := t.getSeatProbility(v.Winner, v.WinType)
		choices = append(choices, utils.Choice{Weight: int(probility * m), Item: v})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found reslut, gameid:%s", userid, t.GameId)
		return false, 0, 0
	}

	// 最终位置
	reslut := choice.Item.(DrawResult)
	return true, reslut.Winner, reslut.WinType
}

func (t *Desk) strategySettlement(userid string) {
	score := t.RBScoreMap[userid]

	for _, s := range t.RBStrategys {
		switch s.Id {
		case JCFS:
			t.JCFSSettlement(userid)
		case HYDT:
			t.HYDTSettlement(userid)
		}

		t.eventPost(userid, event.RB_Strategy, &event.RBStrategySettlementEvent{Id: s.Id, WinScore: score.GetSum()})
	}
}

func (t *Desk) JCFSSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	role.RBStrategy.JCFS.TriggerTimes++
}

func (t *Desk) HYDTSettlement(userid string) {
	role := t.roles[userid]

	if role == nil {
		return
	}

	winScore := t.RBScoreMap[userid]

	strategy := config.GetRBStrategy()
	ctype := handler.GetChargeType(role.User)

	role.RBStrategy.HYDT.WinScore += winScore.GetSum()

	if role.Diamond >= strategy.HYDT.M[ctype] {
		role.RBStrategy.HYDT.U++
		role.RBStrategy.HYDT.WinScore = 0
	}
}
