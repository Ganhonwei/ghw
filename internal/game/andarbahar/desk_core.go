package andarbahar

import (
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
)

// 自然概率开奖
func (t *Desk) NatureWinner() {
	// 指定房间
	// if t.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return
	// }

	var times int
	var poker uint32
	for times, poker = range t.DeskGame.Cards {
		if times%2 == 0 {
			// andar
			t.ACards = append(t.ACards, poker)
		} else {
			// bahar
			t.BCards = append(t.BCards, poker)
		}
		if algo.ABKeyPoker(t.ABDeskFree.Joker, poker) {
			// 这是Key牌
			break
		}
	}

	// 计算谁赢
	t.ABDeskFree.Winner, t.ABDeskFree.SideWinner = algo.ABTimesConvertWinner(times + 1)

	// 计算开牌时间q
	t.ABDeskFree.DrawTime = (len(t.ACards) + len(t.BCards)) + 2
}

// 计算不能开的位置
func (t *Desk) MustLoseSeat() {
	if t.CashBets <= 0 {
		return
	}

	rate := t.Game.AB.RTPFixRate
	m := math.Pow10(6)

	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return
	}

	// 不能开的位置
	noWinSeat := make(map[uint32]int, 0)
	for k, betMap := range t.ABSeatRoleBets {
		if _, ok := betMap[userid]; ok && utils.RandInt64N(int64(m)) <= int64(rate[k-1]*m) {
			noWinSeat[k] = 0
		}
	}

	if len(noWinSeat) == 0 {
		return
	}

	if _, ok := noWinSeat[Andar]; ok {
		if _, ok := noWinSeat[Bahar]; ok {
			// ab同时不能开
			glog.Warningf("andar and bahar both cannot open, gameid:%s", t.GameId)
			return
		}
	}

	if _, ok := noWinSeat[Andar]; !ok {
		if _, ok := noWinSeat[Bahar]; !ok {
			if len(noWinSeat) == 8 {
				// 8个边注的位置不能开
				glog.Warningf("all 8 seats cannot open")
				return
			}
		}
	}

	glog.Infof("gameid:%s,user:%s,noWinSeat:%v", t.GameId, userid, noWinSeat)

	// 把非key牌放到这些位置
	for seat := range noWinSeat {
		pokerIndex := t.getPokerIntervalBySeat(seat)
		for _, index := range pokerIndex {
			if _, ok := t.ABFixedPoker[index]; ok {
				continue
			}
			for i := 0; i < 51; i++ {
				c := t.DeskGame.Cards[0]
				t.DeskGame.Cards = t.DeskGame.Cards[1:]
				if algo.ABKeyPoker(t.ABDeskFree.Joker, c) {
					// 这是Key牌,放回去再重新拿一张
					t.DeskGame.Cards = append(t.DeskGame.Cards, c)
					continue
				}
				t.ABFixedPoker[index] = c
				break
			}
		}
	}

	// 剩下的牌重新洗
	t.shufflePoker()
}

func (t *Desk) StartLead() {
	for i := 0; i < 51; i++ {
		var poker uint32
		if p, ok := t.ABFixedPoker[i+1]; ok {
			poker = p
		} else {
			poker = t.DeskGame.Cards[0]
			t.DeskGame.Cards = t.DeskGame.Cards[1:]
		}

		if i%2 == 0 {
			// andar
			t.ACards = append(t.ACards, poker)
		} else {
			// bahar
			t.BCards = append(t.BCards, poker)
		}

		if algo.ABKeyPoker(t.ABDeskFree.Joker, poker) {
			glog.Infof("gamid:%s, joker:%d, times:%d", t.GameId, t.ABDeskFree.Joker, i)
			return
		}
	}
}

func (t *Desk) SetWinner() {
	times := len(t.ACards) + len(t.BCards)
	t.ABDeskFree.Winner, t.ABDeskFree.SideWinner = algo.ABTimesConvertWinner(times)
	// 计算开牌时间q
	t.ABDeskFree.DrawTime = (len(t.ACards) + len(t.BCards)) + 2
}

func (t *Desk) getSeatProbility(seat uint32) float64 {
	pro := 0.0
	for k, v := range PokerPro {
		switch seat {
		case 1:
			if k%2 != 0 {
				pro += v
			}
		case 2:
			if k%2 == 0 {
				pro += v
			}
		case 3:
			if k >= 1 && k <= 5 {
				pro += v
			}
		case 4:
			if k >= 6 && k <= 10 {
				pro += v
			}
		case 5:
			if k >= 11 && k <= 15 {
				pro += v
			}
		case 6:
			if k >= 16 && k <= 25 {
				pro += v
			}
		case 7:
			if k >= 26 && k <= 30 {
				pro += v
			}
		case 8:
			if k >= 31 && k <= 35 {
				pro += v
			}
		case 9:
			if k >= 36 && k <= 40 {
				pro += v
			}
		default:
			if k >= 41 {
				pro += v
			}
		}
	}

	return pro
}
