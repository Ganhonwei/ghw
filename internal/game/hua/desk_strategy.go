package hua

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"goserver/gen/tb"
	"math"
	"math/rand"
	"sort"
)

const (
	Strategy1001 = "1001" // 怦然心动
	Strategy1002 = "1002" // 乐极生悲
	Strategy1003 = "1003" // 高潮涌现
)

func (t *Desk) checkStrategyOpen(player *data.User, strategyId string) (strategy *tb.TpTpStrategyRecord, open bool) {
	strategy = table.GetTables().TpStrategyTable.Get(strategyId)
	if strategy == nil {
		glog.Errorf("tp strategy %s not exists", strategyId)
		return
	}
	if !strategy.Open {
		return
	}

	var accountTypeOk, chargeTypeOk bool
	for t, open := range strategy.UserAccountTypes {
		if t == player.RegistArea {
			accountTypeOk = open == 1
			break
		}
	}
	if !accountTypeOk {
		return
	}
	chargeType := handler.GetChargeType(player)
	for t, open := range strategy.UserChargeTypes {
		if t == chargeType {
			chargeTypeOk = open == 1
			break
		}
	}
	if !chargeTypeOk {
		return
	}

	// 检查策略互斥
	for _, s := range table.GetTables().TpStrategyTable.GetDataList() {
		if s.Open && s.StrategyId != strategy.StrategyId {
			if utils.SliceIn(strategy.StrategyId, s.MutexStrategy...) { // 互斥
				// 看互斥策略执行了没
				mutexRun := false
				switch s.StrategyId {
				case Strategy1001:
					mutexRun = t.detail.TpStrategy.PRXD.Active
				case Strategy1002:
					mutexRun = t.detail.TpStrategy.LJSB.Active
				case Strategy1003:
					mutexRun = t.detail.TpStrategy.GCYX.Active
				}
				if mutexRun {
					return
				}
			}
		}
	}

	open = true
	return
}

// 玩家是最大高牌换牌
func (t *Desk) checkStrategy1001() {
	playerId := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerId)
	seatId := t.getSeatid(playerId)
	seat := t.getSeat(seatId)
	if player == nil || seat == nil {
		return
	}

	_, open := t.checkStrategyOpen(player, Strategy1001)
	if !open {
		return
	}

	// 高牌
	if algo.HuaType(seat.Cards) != algo.GaoPai {
		return
	}
	// 场上最大
	if !t.isBiggerCards(seatId, false) {
		return
	}

	chargeType := handler.GetChargeType(player)
	strategy := table.GetTables().TpStrategy1001Table.Get()
	// 连续拿高牌局数
	highCardRounds := strategy.HighCardRounds[chargeType]
	if player.TpHighCardRounds+1 < highCardRounds {
		return
	}
	// 换牌概率
	activeRate := strategy.ActiveRates[chargeType]
	if !utils.RandWan(activeRate) {
		return
	}

	t.detail.TpStrategy.PRXD.Active = true
	t.detail.TpStrategy.PRXD.HighCardRounds = player.TpHighCardRounds

	// 换成不是高牌的牌
	// 将要换的手牌添加回牌堆
	deskCards := append(t.DeskGame.Cards, seat.Cards...)
	var cards, remain = seat.Cards, make([]uint32, len(deskCards))
	copy(remain, deskCards)
shuffle:
	for i := 0; i < 100; i++ {
		for len(remain) >= 3 {
			cards, remain = t.getNextCard(3, remain)
			if algo.HuaType(cards) != algo.GaoPai {
				break shuffle
			}
		}
		// 洗牌再发一轮
		remain = make([]uint32, len(deskCards))
		copy(remain, deskCards)
		for i := len(remain) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			remain[i], remain[j] = remain[j], remain[i]
		}
	}
	// 从牌堆移除换掉的牌
	for _, card := range cards {
		deskCards, _ = algo.DealAssignCard(deskCards, card)
	}
	// zlog.Infof("%s change cards %s: %v to %v", t.GameId, playerId, seat.Cards, cards)
	t.DeskGame.Cards = deskCards
	seat.Cards = cards
}

// 记录心率=总玩局数和重置次数(心跳m)
func (t *Desk) updatePlayerStrategy1001Heartbeat(playerId string) {
	player := t.getPlayer(playerId)
	seat := t.getSeat(t.getSeatid(playerId))
	if player == nil || seat == nil {
		return
	}

	_, open := t.checkStrategyOpen(player, Strategy1001)
	if !open {
		return
	}

	player.TpHeartbeatRound++
	huaType := algo.HuaType(seat.Cards)
	msg := &pb.JHHeartbeatSync{}
	if huaType == algo.GaoPai {
		player.TpHighCardRounds++
		msg.AddRound = true
	} else {
		// 重置heartbeat
		player.TpHighCardRounds = 0
		player.TpHeartbeatResetTimes++
		msg.ResetRound = true
	}
	t.send2userid(playerId, msg)
}

// 1002 乐极生悲策略
func (t *Desk) checkStrategy1002() {
	playerId := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerId)
	playerSeatId := t.getSeatid(playerId)
	playerSeat := t.getSeat(playerSeatId)
	if player == nil || playerSeat == nil {
		return
	}

	chargeType := handler.GetChargeType(player)
	strategy := table.GetTables().TpStrategy1002Table.Get()
	cyCD := strategy.Cy[chargeType] + 1 // cd配置更新
	if player.TpUserLjsbPyCD > 0 {
		if player.TpUserLjsbPyCD > cyCD {
			player.TpUserLjsbPyCD = cyCD
		}
		// 冷却-1
		player.TpUserLjsbPyCD--
		// send sync msg
		msg := &pb.JHUserLjsbPyCDSync{PyCD: player.TpUserLjsbPyCD}
		t.send2userid(playerId, msg)
	}

	_, open := t.checkStrategyOpen(player, Strategy1002)
	if !open {
		return
	}

	var pass int
	// 极乐率br
	money := player.Money
	if money == 0 {
		money = uint32(strategy.DefaultMoney)
	}

	br := (float64(player.CashOut) + float64(player.GetScore())) / float64(money)
	if br >= strategy.Br[chargeType] {
		pass += 1
	}
	// 极乐利bp
	bp := player.CashOut + int32(player.GetScore()) - int32(money)
	if bp >= strategy.Bp[chargeType] {
		pass += 2
	}
	if pass == 0 { // 1.超率极乐 2.超利极乐 3.利率极乐
		return
	}
	t.detail.TpStrategy.LJSB.Active = true
	t.detail.TpStrategy.LJSB.JLType = int8(pass)

	opponent := t.isOpponent() // 冤家局
	// 3.1 被冤率py(非冤家局且玩家不是最大),做冤家局,被冤局就不判断后面的冤大率了。
	var isPy bool
	if !opponent && !t.isBiggerCards(playerSeatId, false) && player.TpUserLjsbPyCD <= 0 {
		if utils.RandWan(strategy.Py[chargeType]) {
			isPy = true
			t.detail.TpStrategy.LJSB.Py = true
			player.TpUserLjsbPyCD = cyCD // 被冤率冷却cy
			// send sync msg
			msg := &pb.JHUserLjsbPyCDSync{PyCD: player.TpUserLjsbPyCD}
			t.send2userid(playerId, msg)

			// 随机2手或以上同花的牌
			var tonghuaCards [][]uint32
			var otherCards [][]uint32
			for _, seat := range t.seats {
				if algo.HuaType(seat.Cards) >= algo.TongHua {
					tonghuaCards = append(tonghuaCards, seat.Cards)
				} else {
					otherCards = append(otherCards, seat.Cards)
				}
			}
			tonghuas := len(tonghuaCards)
			max := int(math.Min(float64(len(t.seats)), 3))
			if tonghuas < max { // 同花+大于3副就不管
				needTonghuas := utils.RandMN(max-tonghuas-1, max-tonghuas)
				if needTonghuas > 0 {
					// 获取n副同花以上牌
					cardss, remain := algo.GetCardsHuaTypeBugger(algo.TonghuaDown, needTonghuas, t.DeskGame.Cards)
					t.DeskGame.Cards = remain
					for _, cards := range cardss {
						tonghuaCards = append(tonghuaCards, cards)
						if len(otherCards) > 0 {
							remain = append(remain, otherCards[0]...)
							otherCards = otherCards[1:]
						}
					}
				}
				// 重新派牌成冤家局
				sort.Slice(tonghuaCards, func(i, j int) bool {
					return algo.HuaCompare(tonghuaCards[i], tonghuaCards[j])
				})
				// 非最大同花给玩家
				notBigger := utils.RandMN(1, len(tonghuaCards))
				playerSeat.Cards = tonghuaCards[notBigger]
				otherCards = append(otherCards, tonghuaCards[0:notBigger]...)
				otherCards = append(otherCards, tonghuaCards[notBigger+1:]...)
				// 打乱
				sort.Slice(otherCards, func(i, j int) bool { return utils.RandBool() })
				for _, seat := range t.seats {
					if seat.Userid == playerId {
						continue
					}
					seat.Cards = otherCards[0]
					otherCards = otherCards[1:]
				}
			}
		}
	}

	// 4.冤杀率ps (冤家局且玩家最大牌)
	opponent = t.isOpponent() // 冤家局
	if opponent && t.isBiggerCards(playerSeatId, false) {
		if utils.RandWan(strategy.Ps[chargeType]) {
			t.detail.TpStrategy.LJSB.Ps = true

			// 与人机冤家牌换牌
			var opponentSeats []uint32
			for seatid, seat := range t.seats {
				if t.isRobot(seatid) && algo.HuaType(seat.Cards) >= algo.TongHua {
					opponentSeats = append(opponentSeats, seatid)
				}
			}
			if len(opponentSeats) > 0 {
				// 换牌
				toSeatid := opponentSeats[utils.RandIntN(len(opponentSeats))]
				playerSeat.Cards, t.seats[toSeatid].Cards = t.seats[toSeatid].Cards, playerSeat.Cards
			}
		}
	}

	// 5.冤大率pd (冤家局且非被冤局且玩家不是最大)
	opponent = t.isOpponent() // 冤家局
	if opponent && !isPy && !t.isBiggerCards(playerSeatId, false) {
		if utils.RandWan(strategy.Pd[chargeType]) {
			t.detail.TpStrategy.LJSB.Pd = true

			// 提升牌档位 dd
			level := strategy.Dd[chargeType]
			var cardss [][]uint32
			var playerSeatid uint32
			var robotSeatids []uint32
			for seatid, seat := range t.seats {
				if algo.HuaType(seat.Cards) >= algo.TongHua {
					cardss = append(cardss, seat.Cards)
					if t.isRobot(seatid) {
						robotSeatids = append(robotSeatids, seatid)
					} else {
						playerSeatid = seatid
					}
				}
			}
			toCardss, remainCards := algo.GetCardssHuaTypeBiggerN(cardss, t.DeskGame.Cards, int(level))
			t.DeskGame.Cards = remainCards
			sort.Slice(toCardss, func(i, j int) bool {
				return algo.HuaCompare(toCardss[i], toCardss[j])
			})
			sort.Slice(robotSeatids, func(i, j int) bool {
				return utils.RandBool()
			})
			// 最大牌给人机
			t.seats[robotSeatids[0]].Cards = toCardss[0]
			// 剩下冤家牌打乱再分配
			robotSeatids[0] = playerSeatid
			toCardss = toCardss[1:]
			sort.Slice(toCardss, func(i, j int) bool {
				return utils.RandBool()
			})
			for i, seatid := range robotSeatids {
				t.seats[seatid].Cards = toCardss[i]
			}
		}
	}

	// 6.冤逃率pt(冤家局且玩家最大)
	opponent = t.isOpponent() // 冤家局
	if opponent && t.isBiggerCards(playerSeatId, false) {
		if utils.RandWan(strategy.Pt[chargeType]) {
			t.detail.TpStrategy.LJSB.Pt = true

			dt := strategy.Dt[chargeType]
			for seatId, seat := range t.seats {
				if seatId == playerSeatId {
					continue
				}
				// 降低冤家牌人机牌档位
				if algo.HuaType(seat.Cards) >= algo.TongHua {
					seat.Cards, t.DeskGame.Cards = algo.GetCardsHuaTypeLessN(seat.Cards, t.DeskGame.Cards, int(dt))
				}
			}
		}
	}

	zlog.Infof("%s tp strategy ljsb active %s: opponent=%v, pycd=%d, %#v", t.GameId, playerId, opponent, player.TpUserLjsbPyCD, t.detail.TpStrategy.LJSB)
}

// 是否冤家局, 人机和玩家有同花以上牌
func (t *Desk) isOpponent() (opponent bool) {
	var rHasColor, pHasColor bool
	for seatid, seat := range t.seats {
		if algo.HuaType(seat.Cards) >= algo.TongHua {
			if t.isRobot(seatid) {
				rHasColor = true
			} else {
				pHasColor = true
			}
		}
	}
	opponent = rHasColor && pHasColor
	return
}

// 6 策略高潮涌现
func (t *Desk) checkStrategy1003() {
	playerId := t.GetOnlyOnePlayer()
	player := t.getPlayer(playerId)
	playerSeatId := t.getSeatid(playerId)
	playerSeat := t.getSeat(playerSeatId)
	if player == nil || playerSeat == nil {
		return
	}

	chargeType := handler.GetChargeType(player)
	_, open := t.checkStrategyOpen(player, Strategy1003)
	if !open {
		return
	}
	defer func() {
		// 同步
		msg := &pb.JHUserGCYXSync{
			TpUserGcyxAdd: true,
			TpUserGcyxCD:  player.TpUserGcyxCD,
			TpUserGcyxHp:  player.TpUserGcyxHp,
		}
		t.send2userid(playerId, msg)
	}()

	strategy := table.GetTables().TpStrategy1003Table.Get()

	sCD := strategy.C[chargeType] // cd配置更新
	if player.TpUserGcyxCD > 0 {
		if player.TpUserGcyxCD > sCD {
			player.TpUserGcyxCD = sCD + 1
		}
		// 冷却-1
		player.TpUserGcyxCD--
		return
	}

	baseHp := strategy.Hp[chargeType]
	if player.TpUserGcyxHp == 0 {
		player.TpUserGcyxHp = baseHp
	}
	// 判断生效概率
	if !utils.RandWan(player.TpUserGcyxHp) {
		player.TpUserGcyxHp += strategy.Hp1[chargeType]
		return
	}

	// 冷却, 概率重置
	player.TpUserGcyxCD = sCD
	player.TpUserGcyxHp = baseHp

	t.detail.TpStrategy.GCYX.Active = true

	// 高潮人机数h
	var robotSeats []*data.DeskSeat // 高潮人机
	for seatid, seat := range t.seats {
		if t.isRobot(seatid) {
			robotSeats = append(robotSeats, seat)
		}
	}
	// 牌型降序排序
	sort.Slice(robotSeats, func(i, j int) bool {
		return algo.HuaCompare(robotSeats[i].Cards, robotSeats[j].Cards)
	})
	robotLimit := strategy.H[chargeType]
	if len(robotSeats) < int(robotLimit) {
		robotLimit = int32(len(robotSeats))
	}
	robotNum := utils.RandMN(1, int(robotLimit))
	t.detail.TpStrategy.GCYX.RobotNum = int32(robotNum)

	var highSeats []*data.DeskSeat
	for i, seat := range robotSeats {
		if i < robotNum {
			highSeats = append(highSeats, seat)
		}
	}
	highSeats = append(highSeats, playerSeat)
	sort.Slice(highSeats, func(i, j int) bool {
		return algo.HuaCompare(highSeats[i].Cards, highSeats[j].Cards)
	})

	// 玩家和人机牌型提升 g 档
	level := strategy.G[chargeType]
	var cardss [][]uint32
	for _, seat := range highSeats {
		cardss = append(cardss, seat.Cards)
	}
	toCardss, remainCards := algo.GetCardssHuaTypeBiggerN(cardss, t.DeskGame.Cards, int(level))
	t.DeskGame.Cards = remainCards
	// 提升后大小顺序不变
	sort.Slice(toCardss, func(i, j int) bool {
		return algo.HuaCompare(toCardss[i], toCardss[j])
	})
	for i, seat := range highSeats {
		seat.Cards = toCardss[i]
	}

	isPlayerMax := !t.getPlayer(highSeats[0].Userid).Robot
	// 玩家高潮局返奖率
	var rebateRate float64 = -1
	if player.TpUserGcyxLose != 0 {
		rebateRate = float64(player.TpUserGcyxWin) / float64(-player.TpUserGcyxLose)
	}
	zlog.Infof("%s player 1003 level %d, rebateRate: %.6f", playerId, level, rebateRate)
	if isPlayerMax {
		// 如果玩家最大, 压制率
		if rebateRate > strategy.R[chargeType] {
			if utils.RandWan(strategy.Rp[chargeType]) {
				// 最大牌给人机
				highSeats[0].Cards, highSeats[1].Cards = highSeats[1].Cards, highSeats[0].Cards
				t.detail.TpStrategy.GCYX.Rp = true
			}
		}
	} else {
		// 如果玩家不是最大, 判断恩赐率
		if rebateRate > 0 && rebateRate < strategy.B[chargeType] {
			if utils.RandWan(strategy.Bp[chargeType]) {
				// 最大牌给玩家
				playerSeat.Cards, highSeats[0].Cards = highSeats[0].Cards, playerSeat.Cards
				t.detail.TpStrategy.GCYX.Bp = true
			}
		}
	}
}

// 1003 高潮涌现输赢分记录
func (t *Desk) updatePlayerStrategy1003Score(win bool, userid string, score int64) {
	if !t.detail.TpStrategy.GCYX.Active {
		return
	}
	player := t.getPlayer(userid)
	if player == nil {
		return
	}

	msg := &pb.JHUserGCYXSync{}
	if win {
		player.TpUserGcyxWin += score
		msg.TpUserGcyxWin = score
	} else {
		player.TpUserGcyxLose += score
		msg.TpUserGcyxLose = score
	}
	// 同步
	t.send2userid(userid, msg)
}
