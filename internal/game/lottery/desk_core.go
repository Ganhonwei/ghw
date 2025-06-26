package lottery

import (
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
)

// 自然概率开奖
func (t *Desk) NatureWinner() uint32 {
	// 指定房间
	// if t.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return 1
	// }

	// 开奖权重
	draws := t.Game.LOTTERY.DrawPro

	index, _ := utils.ChoiceIntIndex(draws)
	// 中奖上限
	t.LDetail.IsRandom = true
	return uint32(index + 1)
}

// 计算不能开的位置
func (t *Desk) MustLoseSeat() map[uint32]int {
	// 不能开的位置
	noWinSeat := make(map[uint32]int, 0)
	if t.CashBets <= 0 {
		return noWinSeat
	}

	rate := t.Game.LOTTERY.RTPFixRate
	m := math.Pow10(6)

	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return noWinSeat
	}

	for k, betMap := range t.LSeatRoleBets {
		if _, ok := betMap[userid]; ok && utils.RandInt64N(int64(m)) <= int64(rate[k-1]*m) {
			noWinSeat[k] = 0
		}
	}

	glog.Infof("gameid:%s,user:%s,noWinSeat:%v", t.GameId, userid, noWinSeat)
	return noWinSeat
}

func (t *Desk) StartLead() {
	t.LCards = t.DeskGame.Cards[:3]
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	cardType := algo.HuaType(t.LCards)
	if betmap, ok := t.LSeatRoleBets[cardType]; ok {
		if _, ok := betmap[userid]; ok {
			// 调整
			m := math.Pow10(6)
			rate := t.Game.LOTTERY.RTPFixRate[cardType-1]
			if utils.RandInt64N(int64(m)) <= int64(rate*m) {
				// 不能开
				glog.Infof("userid:%s,trigger fixed rate, seat:%d,gameid:%s", userid, cardType, t.GameId)
				t.UserMustLose()
				return
			}
		}
	}

	// m := math.Pow10(6)
	// canopen := make([]uint32, 0)
	// for seat, betmap := range t.LSeatRoleBets {
	// 	if _, ok := betmap[userid]; ok {
	// 		// 调整
	// 		rate := t.Game.LOTTERY.RTPFixRate[seat-1]
	// 		if utils.RandInt64N(int64(m)) <= int64(rate*m) {
	// 			// 不能开
	// 			glog.Infof("userid:%s,trigger fixed rate, seat:%d,gameid:%s", userid, seat, t.GameId)
	// 			continue
	// 		}
	// 	}
	// 	canopen = append(canopen, seat)
	// }

	// // 当前牌型
	// if utils.SliceIn(cardType, canopen...) {
	// 	return
	// }

	// // 开其他牌型
	// t.ControlLead(canopen)
}

func (t *Desk) ControlLead(seats []uint32) {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	t.shuffle()

	if len(seats) == 0 {
		// 每个位置都下了注,开输得多的
		ctype := t.getUserLoseMostSeat(userid)
		t.LCards, t.DeskGame.Cards = algo.GetCard(ctype, t.DeskGame.Cards)
		return
	}

	if len(seats) == 1 {
		// 只有一个位置可以开，直接开
		t.LCards, t.DeskGame.Cards = algo.GetCard(seats[0], t.DeskGame.Cards)
		return
	}

	// 开指定的牌型
	choices := make([]utils.Choice, 0)
	for _, seat := range seats {
		choices = append(choices, utils.Choice{Weight: int(SeatProMap[seat]), Item: seat})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found losemost reslut, gameid:%s", userid, t.GameId)
		t.LCards = t.DeskGame.Cards[:3]
		return
	}
	t.LCards, t.DeskGame.Cards = algo.GetCard(choice.Item.(uint32), t.DeskGame.Cards)
}

func (t *Desk) UserMustLose() {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	t.shuffle()

	canOpen := make([]uint32, 0)
	for k, betmap := range t.LSeatRoleBets {
		if _, ok := betmap[userid]; !ok {
			// 可以开
			canOpen = append(canOpen, k)
		}
	}

	if len(canOpen) == 0 {
		// 每个位置都下了注,开输得多的
		ctype := t.getUserLoseMostSeat(userid)
		t.LCards, t.DeskGame.Cards = algo.GetCard(ctype, t.DeskGame.Cards)
		return
	}

	if len(canOpen) == 1 {
		// 只有一个位置可以开，直接开
		t.LCards, t.DeskGame.Cards = algo.GetCard(canOpen[0], t.DeskGame.Cards)
		return
	}

	// 开指定的牌型
	choices := make([]utils.Choice, 0)
	for _, seat := range canOpen {
		choices = append(choices, utils.Choice{Weight: int(SeatProMap[seat]), Item: seat})
	}

	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		glog.Warningf("userid:%s, no found losemost reslut, gameid:%s", userid, t.GameId)
		t.LCards = t.DeskGame.Cards[:3]
		return
	}
	t.LCards, t.DeskGame.Cards = algo.GetCard(choice.Item.(uint32), t.DeskGame.Cards)
}
