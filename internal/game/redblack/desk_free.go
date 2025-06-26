package redblack

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"
)

// '初始化
func (t *Desk) freeInit() {
	// 初始化配置的时候可能还没同步过来, 确保有数据
	// for {
	// 	games := config.GetGames()
	// 	if len(games) > 0 {
	// 		glog.Infof("init 红黑 id:%s", t.DeskData.Game.Id)
	// 		t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
	// 		// glog.Infof("init lhd data:%#v", games)
	// 		break
	// 	}
	// }

	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//下注时间
	t.BetTime = int(table.GetTables().RbRoomBaseTable.Get().BetTime)
	//赠送金库存
	t.DeskFree.GiveStock = 0
	t.DeskFree.Cards = make(map[uint32][]uint32) //手牌
	t.DeskFree.Power = make(map[uint32]uint32)   //牌力
	//userid:num, 玩家下注金额
	t.DeskFree.Bets = make(map[string]int64)
	//seat:num, 位置下注金额
	t.DeskFree.SeatBets = make(map[uint32]int64)
	//seat:num, 位置彩金下注金额
	t.DeskFree.SeatCashBets = make(map[uint32]int64)
	//结果 seat:num,seat=(1,2,3,4,5),倍数
	t.DeskFree.Multiple = make(map[uint32]int64)
	//位置(1-5)输赢总量
	t.DeskFree.Score1 = make(map[uint32]int64)
	//每个闲家输赢总量
	t.DeskFree.Score2 = make(map[string]int64)
	t.BackOutDiamond = make(map[string]int64)
	//位置(1-5)上每个玩家输赢
	t.DeskFree.Score3 = make(map[uint32]map[string]int64)
	t.DeskFree.RBDeskFree = &data.RBDeskFree{
		//位置下注详细
		RBSeatRoleBets: make(map[uint32]map[string]data.Currency),
		//人机下注位置
		RBRobotSeat: make(map[string]uint32),
		RBRobotBets: make(map[string]int64),
		// 输赢分
		RBScoreMap: make(map[string]data.Currency),
	}
	//玩家个人系数
	t.DeskFree.UserFactorMap = make(map[string]float64)
	t.state = int32(pb.STATE_READY)
	//赔率
	t.initOdds()
}

//.

// '进入房间响应消息
func (t *Desk) freeEnterMsg(userid string) *pb.RBFreeEnterRoomRsp {
	msg := new(pb.RBFreeEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackRBFreeRoom(t.DeskData)
	//坐下玩家信息
	msg.Userinfo = t.freeSeatBetsMsg()
	t.freeRoomDataMsg(msg.Roominfo)
	return msg
}

func (t *Desk) freeRoomDataMsg(msg *pb.RBFreeRoom) {
	msg.State = t.state
	// 输赢记录20条
	msg.Winer = t.getHistoryMsg()
	tt := 0
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		switch t.state {
		case int32(pb.STATE_READY):
			tt = 1 - t.timer
		case int32(pb.STATE_DEALING):
			tt = FreeDealTime - t.timer
		case int32(pb.STATE_BET):
			tt = t.BetTime - t.timer
		case int32(pb.STATE_LEAD):
			tt = FreeLeadTime - t.timer
			// 开牌阶段要把开的牌发下去
			msg.Poker = t.LeadMsg()
		case int32(pb.STATE_OVER):
			tt = FreeSettlementTime - t.timer
			// 结算阶段要把开的牌发下去
			msg.Poker = t.LeadMsg()
		default:
			tt = t.BetTime - t.timer
		}
	}
	if tt < 0 {
		tt = 0
	}
	msg.Timer = uint32(tt)
}

func (t *Desk) getHistoryMsg() (msg []*pb.RBHistory) {
	for _, h := range t.RBHistory {
		msg = append(msg, &pb.RBHistory{
			Winer: int32(h.Winner),
			Ctype: int32(h.CardType),
		})
	}
	return
}

// 进入消息
func (t *Desk) freeCameinMsg(userid string) {
	msg := new(pb.RBFreeCameinNtf)
	msg.Userinfo = t.freeSeatRoleMsg(userid)
	t.broadcast4(msg)
}

// 位置上玩家数据
func (t *Desk) freeSeatRoleMsg(userid string) (msg *pb.RBFreeUser) {
	var user *data.User
	v := t.roles[userid]
	if v == nil {
		user = t.NewbiewRobot[userid]
	} else {
		user = v.User
	}
	msg = handler.PackRBFreeUser(user)
	if t.DeskFree == nil {
		return
	}
	// 胜场
	if w, ok := user.FreeWinMap[int32(pb.REDBLACK)]; ok {
		for _, v2 := range w {
			if v2.Wtype == 1 {
				msg.Wins++
			}
		}
	}
	msg.Bets = t.userSeatBetMsg(userid)
	// if v, ok := t.roles[userid]; ok {
	// 	// if v.Seat == 0 {
	// 	// 	return //没有坐下不广播
	// 	// }

	// }
	return
}

// 所有坐下玩家数据
func (t *Desk) freeSeatBetsMsg() (msg []*pb.RBFreeUser) {
	for _, v := range t.roles {
		msg2 := t.freeSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	for k, u := range t.NewbiewRobot {
		usermsg := handler.PackRBFreeUser(u)
		if t.DeskFree == nil {
			continue
		}
		// 胜场
		if w, ok := u.FreeWinMap[int32(pb.REDBLACK)]; ok {
			for _, v2 := range w {
				if v2.Wtype == 1 {
					usermsg.Wins++
				}
			}
		}
		usermsg.Bets = t.userSeatBetMsg(k)
		msg = append(msg, usermsg)
	}
	return
}

// '百人下注
func (t *Desk) freeBet(userid string, seatBet uint32,
	num int64) pb.ErrCode {
	if t.state != int32(pb.STATE_BET) {
		return pb.BetOver
	}
	if num <= 0 {
		return pb.OperateError
	}
	if t.Game.Status == 0 {
		return pb.RoomMaintenance
	}
	// if !t.isFree() {
	// 	return pb.NotDealerRoom
	// }
	user := t.getPlayer(userid)
	if user == nil {
		glog.Errorf("userid %s not exist", userid)
		return pb.NotInRoom
	}

	// 充值500以上才能下注
	// if user.Money < 50000 && !user.Robot && !user.SimRobot {
	// 	return pb.FreeBetRecharge
	// }

	role := t.roles[userid]
	if !t.canBet(role, num) {
		return pb.NotEnoughCoin
	}

	base := table.GetTables().RbRoomBaseTable.Get()
	if t.DeskFree.RBSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.RBSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		switch seatBet {
		case 0:
			if limit > int64(base.BetLimit[2]) {
				return pb.BetTopLimit //幸运一击下注限制
			}
		case 1:
			if limit > int64(base.BetLimit[0]) {
				return pb.BetTopLimit //红下注限制
			}
		case 2:
			if limit > int64(base.BetLimit[1]) {
				return pb.BetTopLimit //黑下注限制
			}
		default:
			return pb.OperateError
		}
	}

	if seatBet == RED {
		if _, ok := t.RBSeatRoleBets[BLACK][userid]; ok {
			return pb.BetSeatIllegal
		}
	} else if seatBet == BLACK {
		if _, ok := t.RBSeatRoleBets[RED][userid]; ok {
			return pb.BetSeatIllegal
		}
	}

	cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE106), fmt.Sprintf("红黑%s房间下注", t.DeskData.Rid))
	//下注记录
	t.DeskFree.Bets[userid] += num              //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num         //当前位置总下注额(没扣税的)
	t.DeskFree.SeatCashBets[seatBet] += cashBet //当前位置彩金总下注额
	if !user.GetRobot() {
		t.DeskGame.CashBets += cashBet //当局彩金总下注额
		t.DeskGame.CoinBets += coinBet //当局奖励金总下注额
		if user.OutDiamond-user.Diamond > 0 {
			t.BackOutDiamond[userid] = user.OutDiamond - user.Diamond
		}
	}
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.RBSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Diamond += cashBet
		c.Coin += coinBet
		m[userid] = c
		t.RBSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Diamond: cashBet,
			Coin:    coinBet,
		}
		t.RBSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	role.NoBetTimes = 0 // 不下注次数清零
	msg, ntf := resFreeBet(seatBet, cashBet+coinBet,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.send2userid(userid, msg)
	t.broadcast2(userid, ntf)
	return pb.OK
}

// '新手百人下注
func (t *Desk) newBiewFreeBet(userid string, seatBet uint32,
	num int64) pb.ErrCode {
	if t.state != int32(pb.STATE_BET) {
		return pb.BetOver
	}
	if t.Game.Status == 0 {
		return pb.RoomMaintenance
	}
	user := t.NewbiewRobot[userid]
	if user == nil {
		glog.Errorf("newbiew robot userid %s not exist", userid)
		return pb.NotInRoom
	}

	base := table.GetTables().RbRoomBaseTable.Get()
	if t.DeskFree.RBSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.RBSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		switch seatBet {
		case 0:
			if limit > int64(base.BetLimit[2]) {
				return pb.BetTopLimit //幸运一击下注限制
			}
		case 1:
			if limit > int64(base.BetLimit[0]) {
				return pb.BetTopLimit //红下注限制
			}
		case 2:
			if limit > int64(base.BetLimit[1]) {
				return pb.BetTopLimit //黑下注限制
			}
		default:
			return pb.OperateError
		}
	}
	user.AddCoin(-num)
	t.DeskFree.Bets[userid] += num      //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num //当前位置总下注额(没扣税的)
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.RBSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Coin += num
		m[userid] = c
		t.RBSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Coin: num,
		}
		t.RBSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	_, ntf := resFreeBet(seatBet, num,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.broadcast(ntf)
	return pb.OK
}

// 下注消息
func resFreeBet(beseat uint32, val, coin,
	bets int64, userid string) (*pb.RBFreeBetRsp, *pb.RBFreeBetNtf) {
	rsp := &pb.RBFreeBetRsp{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	ntf := &pb.RBFreeBetNtf{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	return rsp, ntf
}

// 结束重置
func (t *Desk) freeOverInit() {
	// t.freeInit()
	t.state = int32(pb.STATE_READY) //休息停顿
}

//.

// '洗牌
func (t *Desk) shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := make([]uint32, algo.NumCard, algo.NumCard)
	copy(d, algo.NiuCARDS)
	//测试暂时去掉洗牌
	for i := range d {
		j := r.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	t.DeskGame.Cards = d
}

//.

// '发牌
func (t *Desk) freeDeal() {
	t.DeskFree.RBCards = make([][]uint32, 2)
	Red := make([]uint32, 0)
	Black := make([]uint32, 0)
	for i, card := range t.DeskGame.Cards {
		if i%2 == 0 {
			Red = append(Red, card)
		} else {
			Black = append(Black, card)
		}

		if len(Red) == 3 && len(Black) == 3 {
			break
		}
	}

	t.DeskFree.RBCards[0] = Red
	t.DeskFree.RBCards[1] = Black
	t.DeskGame.Cards = t.DeskGame.Cards[6:]
}

//.

// '结束游戏
func (t *Desk) freeGameOver() {

	//赔付
	details := t.gameSettlement()
	//对局详情
	t.gameDetail(details)
	//打印信息
	t.printOver()
	//结束消息
	msg := t.resOverFree()
	//广播
	t.broadcast4(msg)
	//机器人操作
	t.kickRobot()
	//踢除离线玩家
	t.kickOffline()
	//剔除点控房间多余玩家
	t.kickPCMoreUser()
	//奖池清零
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//赠送金库存变化
	if t.GiveStock > 0 {
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "90033")
	}
}

//.

// .结账
func (t *Desk) gameSettlement() []data.RBUserDetail {
	details := make([]data.RBUserDetail, 0)
	var cashStock, coinStock, mcashTaxStock, acashTaxStock, mcoinTaxStock, acoinTaxStock int64 = 0, 0, 0, 0, 0, 0 // 库存变化
	winner, winType := t.getWinnerSeat()
	luck := winType >= algo.LuckDuiZi   // 幸运一击
	odds := t.DeskFree.Multiple[winner] // 赔率

	roleBets := t.DeskFree.RBSeatRoleBets
	scoreMap := make(map[string]data.Currency)
	gamsettelMap := make(map[string]data.Currency) // 结算的钱

	for k, roles := range roleBets {
		// 赢的玩家发奖
		for id, v := range roles {
			// var cash, coin, mcashTax, mcoinTax int64 = 0, 0, 0, 0
			role := t.roles[id]
			if role == nil {
				if !t.robotSettlement(id, k, v) {
					// 新手状态人机结算
					glog.Errorf("redblack Settlement role is nil, id:%s", id)
				}
				continue
			}

			bj := v.Scale(-1)
			var winBets data.Currency
			// 幸运一击
			if k == LUCK && luck {
				odds := t.DeskFree.Multiple[winType]
				winBets.Merge(v.Scale(float64(odds)))
				// bj.Merge(b.Scale(-1))
			}
			// 中奖位置
			if k == winner {
				w := v.Scale(float64(odds))
				winBets.Merge(w)
			}

			// 返奖
			if winBets.GetSum() > 0 {
				if c, ok := gamsettelMap[id]; ok {
					c.Merge(winBets)
					gamsettelMap[id] = c
				} else {
					gamsettelMap[id] = winBets
				}
			}

			winBets.Merge(bj)
			if c, ok := scoreMap[id]; ok {
				c.Merge(winBets)
				scoreMap[id] = c
			} else {
				scoreMap[id] = winBets
			}
		}
	}

	// 结算
	for userid, c := range gamsettelMap {
		t.sendCurrency(userid, c.Coin, c.Diamond, int32(pb.LOG_TYPE107), fmt.Sprintf("红黑房间%s结算", t.Game.Id))
	}

	for k, c := range scoreMap {
		if _, ok := t.roles[k]; !ok {
			// 假人不处理
			continue
		}
		role := t.getUser(k)
		// bet := t.getAllBet(k)
		var mcash, mcoin, realCash int64 = 0, 0, c.Diamond
		var acash, acoin int64
		if c.GetSum() > 0 {
			// 赢了钱,扣税
			mcash, mcoin = t.getTax(realCash, c.Coin, true)
			// 赢了钱,扣税
			acash, acoin = t.getTax(realCash-mcash, c.Coin-mcoin, false)
			// t.sendCurrency(k, -mcoin, -mcash, int32(pb.LOG_TYPE109), fmt.Sprintf("红黑房间%s扣税", t.Game.Id))
		} else {
			// 扣暗税
			realCash = -int64(math.Max(float64(-realCash-role.GiveDiamond), 0))
			acash, acoin = t.getTax(realCash, c.Coin, false)
		}
		// acoin := t.getTax(c.Coin-mcoin, false)
		// 玩家输赢记录
		t.rbRecord(k, c.GetSum())
		// 非机器人记录
		if !role.GetRobot() {
			mcashTaxStock += mcash
			mcoinTaxStock += mcoin
			acashTaxStock += acash
			acoinTaxStock += acoin
			cashStock += (realCash + acash - mcash)
			coinStock += (c.Coin + acoin - mcoin)
			// 事件(非机器人)
			bean := &event.GameRecordEvent{
				Gtype: uint32(pb.REDBLACK),
				Win:   c.GetSum() > 0,
			}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType: int32(pb.REDBLACK),
				Win:      c.GetSum() > 0,
			})

			// 对局详情
			details = append(details, t.createDetail(k, c, 0, 0, acash, acoin)) // mcash, mcoin, acash, acoin
		}
		// 记录输赢分
		// c.Merge(data.Currency{Diamond: -mcash, Coin: -mcoin})
		t.RBScoreMap[k] = c

		// 策略结算
		if !role.Robot {
			t.strategySettlement(role.Userid)
		}
	}
	for _, v := range t.RBObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			details = append(details, t.createDetail(v, data.Currency{}, 0, 0, 0, 0))
		}
	}

	// 游戏开奖记录
	t.addHistory(winner, winType)
	// 库存变化
	t.changeStock(-cashStock, -coinStock, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock, t.Game.Id)
	return details
}

// 输赢记录
func (t *Desk) rbRecord(useid string, num int64) {
	bets := t.DeskFree.Bets[useid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.REDBLACK), Score: num, Bet: bets}
	role := t.roles[useid]
	var result int32 = 1
	if num > 0 {
		// 赢了
		msg.Rtype = 1
	} else if num == 0 {
		msg.Rtype = 0
		result = 0
	} else {
		msg.Rtype = -1
		result = -1
	}
	msg.GameTime = utils.BsonNow().Unix() - t.DeskGame.BeginTime
	t.send2userid(useid, msg)

	wins := make([]data.FreeWin, 0)
	if role.FreeWinMap == nil {
		role.FreeWinMap = make(map[int32][]data.FreeWin)
	}
	if w, ok := role.FreeWinMap[int32(pb.REDBLACK)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.REDBLACK)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	// 游戏局数
	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.REDBLACK)]++

	// 下注
	role.RBStrategy.RoundBet = append(role.RBStrategy.RoundBet, bets)
	if len(role.RBStrategy.RoundBet) > 50 {
		role.RBStrategy.RoundBet = role.RBStrategy.RoundBet[len(role.RBStrategy.RoundBet)-50:]
	}

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    role.Pid,
			Userid:     role.Userid,
			Gtype:      int32(pb.REDBLACK),
			Ts:         time.Now().Unix(),
			Bets:       bets,
			Robot:      role.Robot,
			Username:   role.Nickname,
			Photo:      role.Photo,
			RegistArea: int32(role.RegistArea),
			VipLv:      int32(role.Vip.Lv),
			SuperId:    role.ShareSuperior,
			WaterId:    t.GameId,
			Score:      num,
		}); err != nil {
			glog.Error("publish user redblack bets error", err)
		}

		log := handler.GameFlowWaterLog(bets, int(pb.REDBLACK), 0, len(role.RechargeTarge), useid)
		myactor.Logger().Tell(log)
	}
}

// 开奖记录
func (t *Desk) addHistory(win, winType uint32) {
	t.DeskFree.RBHistory = append(t.DeskFree.RBHistory, data.RBHistory{Winner: win, CardType: winType})
	size := len(t.DeskFree.RBHistory)
	if size > 20 { // 只保留20条
		t.DeskFree.RBHistory = t.DeskFree.RBHistory[size-20:]
	}
}

// . 对局详情
func (t *Desk) createDetail(userid string, score data.Currency, mcash, mcoin, acash, acoin int64) data.RBUserDetail {
	bet := t.Bets[userid]
	detail := data.RBUserDetail{
		Userid:   userid,
		Win:      score.GetSum(),
		Result:   t.getWinReslut(score.GetSum()),
		Observe:  bet <= 0,
		SeatBets: make(map[string]int64),
	}
	for seat, betmap := range t.RBSeatRoleBets {
		if b, ok := betmap[userid]; ok {
			detail.SeatBets[utils.String(seat)] = b.GetSum()
		}
	}

	role := t.roles[userid]
	if role == nil {
		return detail
	}

	detail.AfterScore = role.GetScore()
	detail.AfterCash = role.Diamond
	if score.GetSum() >= 0 {
		detail.BeforeScore = role.GetScore() - score.GetSum() + mcash
		detail.BeforeCash = role.Diamond - score.GetSum() + mcash
	} else {
		detail.BeforeScore = role.GetScore() + bet
		detail.BeforeCash = role.Diamond + bet
	}
	detail.CashMingTax = mcash
	detail.BonusMingTax = mcoin
	detail.CashAnTax = acash
	detail.BonusAnTax = acoin
	return detail
}

//.

// ' 游戏结束消息
func (t *Desk) resOverFree() *pb.RBFreeGameoverNtf {
	winner, winType := t.getWinnerSeat()
	msg := &pb.RBFreeGameoverNtf{
		State:    t.state,
		Userinfo: t.freeSeatBetsMsg(),
		Ctype:    int32(winType),
	}
	msg.Winner = append(msg.Winner, int32(winner))
	if winType >= algo.LuckDuiZi {
		msg.Winner = append(msg.Winner, int32(LUCK))
	}

	for k, v := range t.RBScoreMap {
		bean := &pb.RBRoomScore{
			Userid: k,
		}
		bean.Score = v.GetSum()
		msg.Scores = append(msg.Scores, bean)
		// 跑马灯检测
		t.checkMarquee(k, v)
		// 赠送金变化
		t.checkGiveDiamond(k, v)
		// 分享活动
		t.shareAmount(k, v.Diamond)
		// 时长记录
		t.GameTime(k)
		// 打码量
		t.flowWater(k, v.Diamond)
		// 点控房间
		if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
			t.controlSettlement(k, v.Diamond)
		}
	}
	return msg
}

// 检查桌子状态
func (t *Desk) checkDeskStatus() {
	if t.tickStop {
		return
	}
	t.timer++
	//看看下注时间到没有
	switch t.state {
	case int32(pb.STATE_READY):
		// 空闲状态
		if t.timer >= 1 {
			t.timer = 0
			t.state = int32(pb.STATE_DEALING) // 发牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_DEALING)
			msg.During = FreeDealTime
			t.selfPid.Tell(msg)
			return
		}
	case int32(pb.STATE_DEALING):
		// 发牌状态
		if t.timer >= FreeDealTime {
			t.timer = 0
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_BET)
			msg.During = int32(t.BetTime)
			t.selfPid.Tell(msg)
			return
		}
	case int32(pb.STATE_BET):
		// 下注状态
		if t.timer >= t.BetTime {
			t.timer = 0
			// t.state = int32(pb.STATE_LEAD) // 开牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_LEAD)
			msg.During = FreeLeadTime
			t.selfPid.Tell(msg)
			t.tickStop = true
			return
		}
	case int32(pb.STATE_LEAD):
		// 开牌状态
		if t.timer >= FreeLeadTime {
			t.timer = 0
			// t.state = int32(pb.STATE_OVER) // 结算状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_OVER)
			msg.During = FreeSettlementTime
			t.selfPid.Tell(msg)
			t.tickStop = true
			return
		}
	case int32(pb.STATE_OVER):
		// 结算状态
		if t.timer >= FreeSettlementTime {
			t.timer = 0
			t.state = int32(pb.STATE_READY) // 空闲状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_READY)
			msg.During = 1
			t.selfPid.Tell(msg)
			return
		}
	}
}

// 开牌
func (t *Desk) lead() {
	room, _ := t.getRoom()
	t.RoomId = room
	// 洗牌
	t.shuffle()
	// 发牌
	t.freeDeal()
	// 判断该哪边赢
	t.calWinner()

	// 通知玩家
	ntf := new(pb.RBFreeLeadNtf)
	ntf.Data = t.LeadMsg()
	t.broadcast(ntf)
}

// 根据下注总额获取对应的库存
func (t *Desk) getRoom() (int32, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := table.GetTables().RbRoomBaseTable.Get()
	var roomId int32 = 0
	switch t.Dtype {
	case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
		roomId = games.ControalId
	// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
	// 	roomId = g.RB.NoviceRoom[0]
	case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
		int32(pb.DESK_TYPE_NEWBIEW): // 正常、新手
		intervals := games.BetInterval
		for i, v := range intervals {
			if pool >= int64(v.Nums[0]) && v.Nums[1] == -1 || (pool >= int64(v.Nums[0]) && pool < int64(v.Nums[1])) {
				roomId = games.RoomId[i]
				break
			}
		}
	}

	room := table.GetTables().RbRoomStockTable.Get(roomId)
	if room == nil {
		return 90011, fmt.Errorf("no find room, id:%d", roomId)
	}

	// userid := t.GetOnlyOnePlayer()
	// if userid != "" {
	// 	role := t.roles[userid]
	// 	if role.RegistArea == 0 {
	// 		// A类玩家新手配置
	// 		room.RB.NewbiewMode = room.RB.ANewbiewMode
	// 	}
	// }
	return roomId, nil
}

// 计算哪边赢或者开和
func (t *Desk) calWinner() {
	// 配置结果
	if env == "dev" {
		red, black := t.cfgWinner()
		if red >= algo.GaoPai && red <= algo.LuckBaoZi {
			t.RBCards[RED-1], t.DeskGame.Cards = algo.RBGetCard(red, t.DeskGame.Cards)
		}
		if black >= algo.GaoPai && black <= algo.LuckBaoZi {
			t.RBCards[BLACK-1], t.DeskGame.Cards = algo.RBGetCard(black, t.DeskGame.Cards)
		}
		if red != 0 || black != 0 {
			return
		}
	}

	// 选择策略
	t.selectStrategy()
	trigger := false
	var seat, cardType uint32
	for _, s := range t.RBDeskFree.RBStrategys {
		switch s.Id {
		case HYDT:
			trigger, seat, cardType = t.HYDTStrategy()
		case JCFS:
			trigger, seat, cardType = t.JCFSStrategy()
		}
		if trigger {
			t.RBDeskFree.RBTriggerStrategy = s.Id
			// 判断是否是指定牌型
			if seat > 0 {
				t.controalCardType(seat, cardType)
			}
		}
	}

	// 没触发策略或者没有指定牌型，按调整系数走随机发牌
	if !trigger || seat <= 0 {
		// 自然概率
		t.NatureWinner()
	}
}

// 开牌结果msg
func (t *Desk) LeadMsg() *pb.RBFreeRoomOver {
	return &pb.RBFreeRoomOver{
		Red:       t.RBCards[0],
		Redtype:   int32(algo.RedBlackType(t.RBCards[0])),
		Black:     t.RBCards[1],
		Blacktype: int32(algo.RedBlackType(t.RBCards[1])),
	}
}

// 玩家下注总额
func (t *Desk) userSeatBetMsg(userid string) []*pb.RBFreeUserBet {
	bets := make([]*pb.RBFreeUserBet, 0)
	for k, v := range t.DeskFree.RBSeatRoleBets {
		value := v[userid]
		msg := new(pb.RBFreeUserBet)
		msg.Ptype = int32(k)
		msg.Bets = uint32(value.GetSum())
		bets = append(bets, msg)
	}
	return bets
}

// 事件
func (t *Desk) eventPost(userid string, eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("redblack event Marshal fail")
		return
	}
	msg := new(pb.EventPost)
	msg.EventId = eventId
	msg.Data = body
	t.send2userid(userid, msg)
}

// 发牌阶段
func (t *Desk) dealing() {
	// 每个玩家不下注轮数+1
	for _, dr := range t.roles {
		if dr.GetRobot() {
			continue
		}
		dr.NoBetTimes++
	}
}

// 对局详情
func (t *Desk) gameDetail(details []data.RBUserDetail) {
	if len(details) <= 0 {
		// 新手房没下注的不记
		return
	}
	winner, winType := t.getWinnerSeat()
	t.RBDetail.Winner = append(t.RBDetail.Winner, winner)
	if winType >= algo.LuckDuiZi {
		t.RBDetail.Winner = append(t.RBDetail.Winner, LUCK)
	}
	t.RBDetail.Cards = t.RBCards
	t.RBDetail.CardType = []uint32{algo.RedBlackType(t.RBCards[0]), algo.RedBlackType(t.RBCards[1])}
	t.RBDetail.Bets = t.CashBets + t.CoinBets
	t.RBDetail.StrategyId = t.RBDeskFree.RBTriggerStrategy
	t.RBDetail.UserDetail = details

	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	t.RBDetail.PlayerWin = playerWin
	// 保存详情
	t.saveDetail()
}

// 保存详情
func (t *Desk) saveDetail() {
	detail := data.Detail{
		WaterId:      t.GameId,
		BeginTime:    t.BeginTime,
		EndTime:      utils.BsonNow().Unix(),
		Gtype:        int32(pb.REDBLACK),
		RoomId:       t.Rid,
		DeskId:       t.Game.Id,
		PlayerFactor: int(t.DeskGame.FinaFactor * 100),
		RBDetail:     &t.RBDetail,
	}
	players := make([]string, 0)
	for _, ld := range t.RBDetail.UserDetail {
		players = append(players, ld.Userid)
	}
	detail.Players = strings.Join(players, ",")
	body, err := json.Marshal(detail)
	if err != nil {
		glog.Errorf("save detail error %#v", detail)
		return
	}
	msg := &pb.Detail{}
	msg.Data = body
	myactor.Logger().Tell(msg)
}

// 点控房间结算
func (t *Desk) controlSettlement(userid string, cash int64) {
	if role, ok := t.roles[userid]; ok && !role.Robot {
		if !role.PCSwitch {
			return
		}
		role.PCScoreComplete += cash
		if role.PCScore > 0 { //控赢
			if role.PCScoreComplete >= role.PCScore {
				role.PCSwitch = false
				role.PCFactor = 0
				role.PCScoreComplete = 0
				role.PCScore = 0
				t.Dtype = int32(pb.DESK_TYPE_NORMAL)
			}
		} else if role.PCScore < 0 { //控输
			if role.PCScoreComplete <= role.PCScore {
				role.PCSwitch = false
				role.PCFactor = 0
				role.PCScoreComplete = 0
				role.PCScore = 0
				t.Dtype = int32(pb.DESK_TYPE_NORMAL)
			}
		}
		t.sendPointControl(role.Userid)
		return
	}
}

// 新手人机结算
func (t *Desk) robotSettlement(id string, seat uint32, c data.Currency) bool {
	winner, winType := t.getWinnerSeat()
	luck := winType >= algo.LuckDuiZi // 幸运一击

	if user, ok := t.NewbiewRobot[id]; ok {
		var winScore data.Currency
		if seat == winner {
			odds := t.DeskFree.Multiple[winner] // 赔率
			user.AddCoin(c.GetSum() * int64(odds))
			winScore = c.Scale(float64(odds))
			// t.RBScoreMap[id] = c.Scale(float64(odds))
		} else {
			winScore = c.Scale(-1)
		}
		// 幸运一击
		if luck && seat == LUCK {
			if betmap, ok := t.RBSeatRoleBets[seat]; ok {
				if b, ok := betmap[id]; ok {
					odds := t.DeskFree.Multiple[winType] // 赔率
					user.AddCoin(b.GetSum() * int64(odds))
					winScore = c.Scale(float64(odds))
				}
			}
		}

		if s, ok := t.RBScoreMap[id]; ok {
			s.Merge(winScore)
			t.RBScoreMap[id] = s
		} else {
			t.RBScoreMap[id] = winScore
		}
		return true
	}
	return false
}

// 获取唯一玩家
func (t *Desk) GetOnlyOnePlayer() string {
	userids := []string{}
	for k, v := range t.roles {
		if !v.User.GetRobot() {
			userids = append(userids, k)
		}
	}

	if len(userids) == 1 {
		return userids[0]
	} else {
		return ""
	}
}

func (t *Desk) getWinnerSeat() (winSeat uint32, winType uint32) {
	winSeat = BLACK
	winType = 1
	if len(t.RBCards) <= 0 {
		return
	}
	winType = algo.RedBlackType(t.RBCards[1])
	if algo.HuaCompare(t.RBCards[0], t.RBCards[1]) {
		winSeat = RED
		winType = algo.RedBlackType(t.RBCards[0])
	}
	return
}

// vim: set foldmethod=marker foldmarker=//',//.:
