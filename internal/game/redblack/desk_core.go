package redblack

import (
	"goserver/pkg/game/algo"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
)

// 自然概率开奖
func (t *Desk) NatureWinner() {
	// 计算能赢的位置
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	bet := t.Bets[userid]
	if bet <= 0 {
		return
	}

	// winSeat, winType := t.getWinnerSeat()

	room := table.GetTables().RbRoomBaseTable.Get()
	// 每个下注位置输赢
	winScore := make(map[uint32]int64)  // 红黑输赢分
	excludeSeat := make(map[uint32]int) // 不可开的位置
	mustLose := false

	// 红黑输赢分
	m := math.Pow10(8)
	for seat, betmap := range t.RBSeatRoleBets {
		if c, ok := betmap[userid]; ok {
			// 判断不能开的位置
			if seat == LUCK {
				for i := LUCkPAIR; i <= SET; i++ {
					fix := float64(room.FixRate[i-1])
					if utils.RandInt32N(int32(m)) <= int32(m*fix) {
						excludeSeat[LUCK] = 1
						break
					}
				}
			} else {
				fix := float64(room.FixRate[seat-1])
				if utils.RandInt32N(int32(m)) <= int32(m*fix) {
					excludeSeat[seat] = 1
				}
			}
			if seat == LUCK {
				continue
			}
			//输赢分,幸运一击和红黑其中的一门输赢分应该是一样的
			odds := t.DeskFree.Multiple[seat]
			winScore[seat] = c.GetSum()*odds - bet
		} else {
			if seat != LUCK {
				winScore[seat] = -bet
			}
		}
	}

	if len(excludeSeat) == 0 {
		return
	}

	if len(excludeSeat) >= 3 {
		mustLose = true
	}

	if mustLose {
		// 开输得多的
		t.loseMost(userid, winScore)
		return
	}

	// 又不能开的
	_, luckSeat := excludeSeat[LUCK]
	if _, ok := excludeSeat[RED]; ok {
		// 红不能开就开黑
		if luckSeat {
			t.controalDeal(BLACK, false)
		} else {
			// 幸运一击可开可不开
			t.ChangeCards(BLACK)
		}
	} else if _, ok := excludeSeat[BLACK]; ok {
		// 黑不能开就开红
		if luckSeat {
			t.controalDeal(RED, false)
		} else {
			// 幸运一击可开可不开
			t.ChangeCards(RED)
		}
	} else {
		// 红黑都能开， 但是不能开幸运一击
		t.controalDeal(0, false)
	}
}

func (t *Desk) loseMost(userid string, winScore map[uint32]int64) {
	loseSeat := make([]uint32, 0) // 这里面应该只有红或者黑

	// 幸运一击
	var luckSeat bool
	if b, ok := t.RBSeatRoleBets[LUCK]; ok && b[userid].GetSum() > 0 {
		luckSeat = true
	}

	var loseScore int64
	for seat, score := range winScore {
		if score > 0 || loseScore < score {
			continue
		}
		loseSeat = append(loseSeat, seat)
		loseScore = score
	}

	if len(loseSeat) <= 0 {
		// 没有输的最多的，随机开
		return
	}

	seat, _ := utils.ChoiceUint32(loseSeat)

	if !luckSeat {
		// 幸运一击没下注，直接换牌就行
		t.ChangeCards(seat)
	} else {
		// 幸运一击下了注，不能开
		t.controalDeal(seat, false)
	}
}

func (t *Desk) controalDeal(seat uint32, luckSeat bool) {
	// winSeat, winType := t.getWinnerSeat()
	if !luckSeat {
		// 不能开幸运一击，换牌
		for {
			if len(t.DeskGame.Cards) <= 1 {
				break
			}
			cards := t.DeskGame.Cards[:3]
			t.DeskGame.Cards = t.DeskGame.Cards[1:]
			if algo.RedBlackType(cards) < algo.LuckDuiZi {
				// 保证2副牌都小于对9
				if algo.RedBlackType(t.RBCards[0]) >= algo.LuckDuiZi {
					t.RBCards[0] = cards
					continue
				} else if algo.RedBlackType(t.RBCards[1]) >= algo.LuckDuiZi {
					t.RBCards[1] = cards
					continue
				}
				break
			}
		}
	} else {
		winSeat, winType := t.getWinnerSeat()
		if winType < algo.LuckDuiZi {
			// 要开幸运一击,换牌
			for {
				if len(t.DeskGame.Cards) <= 2 {
					break
				}
				cards := t.DeskGame.Cards[:3]
				t.DeskGame.Cards = t.DeskGame.Cards[1:]
				if algo.RedBlackType(cards) >= algo.LuckDuiZi {
					// 保证1副牌大于等于对9
					t.RBCards[winSeat-1] = cards
					break
				}
			}
		}
	}
	// 换牌，把大的牌给指定的位置
	t.ChangeCards(seat)
	// if algo.HuaCompare(t.RBCards[0], t.RBCards[1]) {
	// 	if seat == BLACK {
	// 		t.RBCards[0], t.RBCards[1] = t.RBCards[1], t.RBCards[0]
	// 	}
	// } else {
	// 	if seat == RED {
	// 		t.RBCards[0], t.RBCards[1] = t.RBCards[1], t.RBCards[0]
	// 	}
	// }
}

func (t *Desk) controalCardType(seat, winType uint32) {
	// winSeat, winType := t.getWinnerSeat()
	if winType < algo.LuckDuiZi {
		// 不能开幸运一击，换牌
		for {
			if len(t.DeskGame.Cards) <= 1 {
				break
			}
			cards := t.DeskGame.Cards[:3]
			t.DeskGame.Cards = t.DeskGame.Cards[1:]
			if algo.RedBlackType(cards) < algo.LuckDuiZi {
				// 保证2副牌都小于对9
				if algo.RedBlackType(t.RBCards[0]) >= algo.LuckDuiZi {
					t.RBCards[0] = cards
					continue
				} else if algo.RedBlackType(t.RBCards[1]) >= algo.LuckDuiZi {
					t.RBCards[1] = cards
					continue
				}
				break
			}
		}
	} else {
		// 要开幸运一击,换牌
		t.RBCards[seat-1], t.DeskGame.Cards = algo.RBGetCard(winType, t.DeskGame.Cards)
		return
	}
	// 换牌，把大的牌给指定的位置
	t.ChangeCards(seat)
}

func (t *Desk) getSeatProbility(seat, CardType uint32) float64 {
	typePro := map[uint32]float64{RED: 0.4737, BLACK: 0.4737, LUCkPAIR: 0.1371, COLOR: 0.0915, SEQ: 0.0628, PURESEQ: 0.0043, SET: 0.0043}

	pro := typePro[seat]
	if CardType > 0 {
		pro = pro * typePro[CardType]
	}

	return pro
}

func (t *Desk) ChangeCards(winSeat uint32) {
	if len(t.RBCards) < 2 {
		return
	}
	if algo.HuaCompare(t.RBCards[0], t.RBCards[1]) {
		if winSeat == BLACK {
			t.RBCards[0], t.RBCards[1] = t.RBCards[1], t.RBCards[0]
		}
	} else {
		if winSeat == RED {
			t.RBCards[0], t.RBCards[1] = t.RBCards[1], t.RBCards[0]
		}
	}
}
