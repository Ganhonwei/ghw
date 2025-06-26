package andarbahar

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"
)

// '初始化
func (t *Desk) freeInit() {
	// 初始化配置的时候可能还没同步过来, 确保有数据
	for {
		games := config.GetGames()
		if len(games) > 0 {
			glog.Infof("init andarbahar id:%s", t.DeskData.Game.Id)
			t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
			break
		}
	}

	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	//下注时间
	t.BetTime = int(t.Game.AB.BetTime)
	//赠送金库存
	t.DeskFree.GiveStock = 0
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
	t.DeskFree.ABDeskFree = &data.ABDeskFree{
		//输赢分
		ABScoreMap: make(map[string]data.Currency),
		OddsMap:    make(map[uint32]int),
		//位置下注详细
		ABSeatRoleBets: make(map[uint32]map[string]data.Currency),
		ABFixedPoker:   make(map[int]uint32),
	}
	//玩家个人系数
	t.DeskFree.UserFactorMap = make(map[string]float64)

	t.state = int32(pb.STATE_READY)
	//赔率
	t.initOdds()
}

//.

// '进入房间响应消息
func (t *Desk) freeEnterMsg(userid string) *pb.ABFreeEnterRoomRsp {
	msg := new(pb.ABFreeEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackABFreeRoom(t.DeskData)
	//坐下玩家信息
	msg.Userinfo = t.freeSeatBetsMsg()
	//排行榜前6玩家
	msg.Rank = t.freeRankMsg(20)
	t.freeRoomDataMsg(msg.Roominfo)
	return msg
}

func (t *Desk) freeRoomDataMsg(msg *pb.ABFreeRoom) {
	msg.State = t.state
	// 输赢记录20条
	msg.History = t.buildHistory()
	msg.Bets = t.seatBetMsg()
	tt := 0
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		switch t.state {
		case int32(pb.STATE_READY):
			tt = ReadyTime - t.timer
		case int32(pb.STATE_BET):
			tt = t.BetTime*2 - t.timer
			msg.Poker = t.LeadMsg()
		case int32(pb.STATE_LEAD):
			tt = t.DrawTime - t.timer
			// 开牌阶段要把开的牌发下去
			msg.Poker = t.LeadMsg()
		case int32(pb.STATE_OVER):
			tt = FreeSettlementTime - t.timer
			// 结算阶段要把开的牌发下去
			msg.Poker = t.LeadMsg()
		default:
			tt = t.BetTime*2 - t.timer
		}
	}
	if tt < 0 {
		tt = 0
	}

	msg.Timer = uint32(tt / 2)
}

// 进入消息
func (t *Desk) freeCameinMsg(userid string) {
	msg := new(pb.ABFreeCameinNtf)
	msg.Userinfo = t.freeSeatRoleMsg(userid)
	t.broadcast4(msg)
}

// 位置上玩家数据
func (t *Desk) freeSeatRoleMsg(userid string) (msg *pb.ABFreeUser) {
	var user *data.User
	v := t.roles[userid]
	if v == nil {
		user = t.NewbiewRobot[userid]
	} else {
		user = v.User
	}
	msg = handler.PackABFreeUser(user)
	if t.DeskFree == nil {
		return
	}
	// 胜场
	if w, ok := user.FreeWinMap[int32(pb.ABAR)]; ok {
		for _, v2 := range w {
			if v2.Wtype == 1 {
				msg.Wins++
			}
		}
	}
	msg.Bets = t.userSeatBetMsg(userid)
	return
}

// 所有坐下玩家数据
func (t *Desk) freeSeatBetsMsg() (msg []*pb.ABFreeUser) {
	for _, v := range t.roles {
		msg2 := t.freeSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	for k, u := range t.NewbiewRobot {
		usermsg := handler.PackABFreeUser(u)
		if t.DeskFree == nil {
			continue
		}
		// 胜场
		if w, ok := u.FreeWinMap[int32(pb.ABAR)]; ok {
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

// 排行榜数据
func (t *Desk) freeRankMsg(num int32) (msg []*pb.ABFreeUser) {
	var users []data.User
	for _, v := range t.roles {
		users = append(users, *v.User)
	}
	for _, u := range t.NewbiewRobot {
		users = append(users, *u)
	}
	sort.Slice(users, func(i, j int) bool { return (users[i].Diamond + users[i].Coin) > (users[j].Diamond + users[j].Coin) })
	// 取前num位玩家
	for _, dr := range users {
		if len(msg) >= int(num) {
			break
		}
		msg1 := t.freeSeatRoleMsg(dr.Userid)
		if msg1 == nil {
			continue
		}
		msg = append(msg, msg1)
	}
	return
}

//.

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
	user := t.getPlayer(userid)
	if user == nil {
		glog.Errorf("userid %s not exist", userid)
		return pb.NotInRoom
	}

	role := t.roles[userid]
	if !t.canBet(role, num) {
		return pb.NotEnoughCoin
	}
	if t.DeskFree.ABSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.ABSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		betLimit := t.DeskData.Game.AB.BetLimit[seatBet-1]
		if limit > int64(betLimit) {
			return pb.BetTopLimit //下注限制
		}
	}
	cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE87), fmt.Sprintf("andarbahar%s房间下注", t.DeskData.Rid))
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
	if m, ok := t.ABSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Diamond += cashBet
		c.Coin += coinBet
		m[userid] = c
		t.ABSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Diamond: cashBet,
			Coin:    coinBet,
		}
		t.ABSeatRoleBets[seatBet] = m
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
	if t.DeskFree.ABSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.ABSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		betLimit := t.DeskData.Game.AB.BetLimit[seatBet-1]
		if limit > int64(betLimit) {
			return pb.BetTopLimit //下注限制
		}
	}
	user.AddDiamond(-num)
	t.DeskFree.Bets[userid] += num      //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num //当前位置总下注额(没扣税的)
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.ABSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Diamond += num
		m[userid] = c
		t.ABSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Diamond: num,
		}
		t.ABSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	_, ntf := resFreeBet(seatBet, num,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.broadcast(ntf)
	return pb.OK
}

// 下注消息
func resFreeBet(beseat uint32, val, coin,
	bets int64, userid string) (*pb.ABFreeBetRsp, *pb.ABFreeBetNtf) {
	rsp := &pb.ABFreeBetRsp{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	ntf := &pb.ABFreeBetNtf{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	return rsp, ntf
}

//.

//.

// ' 超时处理
func (t *Desk) freeTimeout() {
	switch t.state {
	case int32(pb.STATE_READY):
		t.freeStart()
	case int32(pb.STATE_BET):
		if t.timer == FreeBetTime {
			//出牌超时处理
			t.timer = 0
			t.freeGameOver()
		} else {
			t.timer++
		}
	case int32(pb.STATE_OVER):
		if t.timer == RestTime {
			//出牌超时处理
			t.timer = 0
			t.gameStartBet()
		} else {
			t.timer++
		}
	default:
		t.timer++
	}
}

//.

//'游戏状态

// 结束重置
func (t *Desk) freeOverInit() {
	// t.freeInit()
	t.state = int32(pb.STATE_READY) //休息停顿
}

// 开始状态初始化
func (t *Desk) freeStart() {
	if len(t.roles) > 0 {
		t.state = int32(pb.STATE_OVER) //休息停顿
	} else {
		t.state = int32(pb.STATE_READY) //准备
		return
	}
	t.freeStartMsg()
}

// 状态变更消息
func (t *Desk) freeStartMsg() {
	// var photo string
	// var nickname string
	// p := t.getPlayer(t.DeskGame.Dealer)
	// if p != nil {
	// 	photo = p.GetPhoto()
	// 	nickname = p.GetNickname()
	// }
	// var left uint32 = t.leftDealerTimes()
	// msg := resFreeStart(t.DeskGame.Dealer, photo, nickname,
	// 	t.state, t.DeskFree.Carry, DealerTimes, left)
	// t.broadcast(msg)
}

// 开始下注
func (t *Desk) gameStartBet() {
	//下注状态
	t.state = int32(pb.STATE_BET)
	t.freeStartMsg()
}

// 剩余坐庄次数
func (t *Desk) leftDealerTimes() uint32 {
	if DealerTimes >= t.DeskFree.DealerNum {
		return DealerTimes - t.DeskFree.DealerNum
	}
	glog.Errorf("Dealer %s, DealerNum %d", t.DeskGame.Dealer, t.DeskFree.DealerNum)
	return 0
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

func (t *Desk) shufflePoker() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := make([]uint32, len(t.DeskGame.Cards))
	copy(d, t.DeskGame.Cards)
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
	// 计算每种结果的平台输赢分(从大到小排序)
	// draws := t.calWinningScoreOf16()
	// // 最终系数
	// factor := t.getFinalFactor()
	// // 选择牌型
	// cardType := t.Game.AB.CardTypeID[len(t.Game.AB.FinalCoefficient)]
	// for i, v := range t.Game.AB.FinalCoefficient {
	// 	if factor <= float64(v)/100 {
	// 		cardType = t.Game.AB.CardTypeID[i]
	// 		break
	// 	}
	// }
	// glog.Infof("andarbahar factor:%.2f, cardtype:%d,rid:%s,gameId:%s", factor, cardType, t.Rid, t.GameId)

	// card := t.Game.AB.CardType[0]
	// for _, c := range t.Game.AB.CardType {
	// 	if c.Id == int(cardType) {
	// 		card = c
	// 	}
	// }

	// 选择开哪个位置

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
	//踢除离线玩家
	t.kickOffline()
	//剔除5局不下注的玩家
	// t.kickNoBet()
	//剔除点控房间多余玩家
	t.kickPCMoreUser()
	//奖池清零
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//赠送金库存变化
	if t.GiveStock > 0 {
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "60033")
	}
}

//.

// .结账
func (t *Desk) gameSettlement() []*data.ABUserDetail {
	details := make([]*data.ABUserDetail, 0)
	var cashStock, coinStock, mcashTaxStock, acashTaxStock, mcoinTaxStock, acoinTaxStock int64 = 0, 0, 0, 0, 0, 0 // 库存变化
	winner, sideWinner := t.ABDeskFree.Winner, t.ABDeskFree.SideWinner
	// odds := t.DeskFree.ABDeskFree.OddsMap[winner] // 赔率
	roleBets := t.DeskFree.ABSeatRoleBets
	scoreMap := make(map[string]data.Currency)
	for k, roles := range roleBets {
		// 赢的玩家发奖
		for id, v := range roles {
			// var cash, coin, mcashTax, mcoinTax int64 = 0, 0, 0, 0
			role := t.roles[id]
			if role == nil {
				if !t.robotSettlement(id, k, v) {
					// 新手状态人机结算
					glog.Errorf("andarbahar Settlement role is nil, id:%s", id)
				}
				continue
			}

			winBets := v.Scale(-1)
			if winner == k {
				// 押中andar或者bahar
				odds := t.DeskFree.ABDeskFree.OddsMap[winner] // 赔率
				w := v.Scale(float64(odds) / 100)
				winBets.Merge(w)
				// 返奖
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE88), fmt.Sprintf("ab房间%s位置%d赢分", t.Game.Id, k))
			} else if sideWinner == k {
				// 押中边注
				odds := t.DeskFree.ABDeskFree.OddsMap[sideWinner] // 边注赔率
				w := v.Scale(float64(odds) / 100)
				winBets.Merge(w)
				// 返奖
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE88), fmt.Sprintf("ab房间%s位置%d赢分", t.Game.Id, k))
			}

			if c, ok := scoreMap[id]; ok {
				c.Merge(winBets)
				scoreMap[id] = c
			} else {
				scoreMap[id] = winBets
			}
		}
	}
	for k, c := range scoreMap {
		if _, ok := t.roles[k]; !ok {
			// 假人不处理
			continue
		}
		role := t.getUser(k)
		var mcash, mcoin, realCash int64 = 0, 0, c.Diamond
		var acash, acoin int64
		if c.GetSum() > 0 {
			// 赢了钱,扣税
			mcash, mcoin = t.getTax(realCash, c.Coin, true)
			// 赢了钱,扣税
			acash, acoin = t.getTax(realCash-mcash, c.Coin-mcoin, false)
			// 扣明税
			// t.sendCurrency(k, -mcoin, -mcash, int32(pb.LOG_TYPE89), fmt.Sprintf("ab房间%s明税", t.Game.Id))
		} else {
			// 扣暗税
			realCash = -int64(math.Max(float64(-realCash-role.GiveDiamond), 0))
			acash, acoin = t.getTax(realCash, c.Coin, false)
		}
		// acoin := t.getTax(c.Coin-mcoin, false)
		// 玩家输赢记录
		t.abRecord(k, c.GetSum())
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
				Gtype: uint32(pb.ABAR),
				Win:   c.GetSum() > 0,
			}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType: int32(pb.ABAR),
				Win:      c.GetSum() > 0,
			})
			// 对局详情
			details = append(details, t.createDetail(k, c, 0, 0, acash, acoin)) //, mcash, mcoin, acash, acoin))
		}
		// 记录输赢分
		// c.Merge(data.Currency{Diamond: -mcash, Coin: -mcoin})
		t.ABScoreMap[k] = c

		if !role.GetRobot() {
			t.strategySettlement(k)
		}
	}

	for _, v := range t.ABObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			details = append(details, t.createDetail(v, data.Currency{}, 0, 0, 0, 0))
		}
	}

	// 游戏开奖记录
	t.addHistory()
	// 库存变化
	t.changeStock(-cashStock, -cashStock, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock, t.Game.Id)
	return details
}

// 输赢记录
func (t *Desk) abRecord(useid string, num int64) {
	bets := t.DeskFree.Bets[useid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.ABAR), Score: num, Bet: bets}
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
	if w, ok := role.FreeWinMap[int32(pb.ABAR)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.ABAR)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.ABAR)]++

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    role.Pid,
			Userid:     role.Userid,
			Gtype:      int32(pb.ABAR),
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
			glog.Error("publish user abar bets error", err)
		}

		log := handler.GameFlowWaterLog(bets, int(pb.ABAR), 0, len(role.RechargeTarge), useid)
		myactor.Logger().Tell(log)
	}

}

// 开奖记录
func (t *Desk) addHistory() {
	t.ABWinnerSeat = append(t.ABWinnerSeat, t.ABDeskFree.Winner)
	size := len(t.ABWinnerSeat)
	if size > 100 { // 只保留100条
		t.ABWinnerSeat = t.ABWinnerSeat[size-100:]
	}
	// joker牌记录
	t.ABJoker = append(t.ABJoker, t.ABDeskFree.Joker)
	size = len(t.ABJoker)
	if size > 8 { // 只保留8条
		t.ABJoker = t.ABJoker[size-8:]
	}
}

// '记录房间上局赢家
func (t *Desk) saveWiners() {
	t.DeskFree.Winers = make([]*data.FreeWiner, 0)
	for k, v := range t.DeskFree.Score2 {
		if v < 0 {
			continue
		}
		if (v - t.DeskFree.Bets[k]) < 0 {
			continue
		}
		user := t.getPlayer(k)
		if user == nil {
			glog.Errorf("userid %s not exist", k)
			continue
		}
		winer := &data.FreeWiner{
			Userid:   k,
			Nickname: user.GetNickname(),
			Photo:    user.GetPhoto(),
			Coin:     v - t.DeskFree.Bets[k],
		}
		t.DeskFree.Winers = append(t.DeskFree.Winers, winer)
	}
}

// . 对局详情
func (t *Desk) createDetail(userid string, score data.Currency, mcash, mcoin, acash, acoin int64) *data.ABUserDetail {
	bet := t.Bets[userid]
	detail := &data.ABUserDetail{
		Userid:   userid,
		SeatBets: t.getSeatBetsOfUserid(userid),
		Win:      score.GetSum(),
		Result:   t.getWinReslut(score.GetSum()),
		Observe:  bet <= 0,
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
	// detail.AfterScore = role.GetScore()
	// detail.AfterBonus = role.Coin
	// detail.AfterCash = role.Diamond
	// detail.BeforeScore = role.GetScore() - score.GetSum()
	// detail.BeforeBonus = role.Coin - score.Coin
	// detail.BeforeCash = role.Diamond - score.Diamond
	detail.CashMingTax = mcash
	detail.BonusMingTax = mcoin
	detail.CashAnTax = acash
	detail.BonusAnTax = acoin
	return detail
}

// '个人记录
func (t *Desk) setFreeRecord() {
	for k, v := range t.DeskFree.Score2 {
		pid := t.getPid(k)
		if pid == nil {
			continue
		}
		user := t.getPlayer(k)
		if user == nil {
			glog.Errorf("userid %s not exist", k)
			continue
		}
		msg := new(pb.SetRecord)
		if v > 0 {
			msg.Rtype = 1
		} else if v < 0 {
			msg.Rtype = -1
		} else {
			msg.Rtype = 0
		}
		//更新游戏内数据
		user.SetRecord(msg.Rtype)
		//更新节点数据
		pid.Tell(msg)
	}
}

//.

// '庄家收钱,TODO 奖池抽成
func (t *Desk) dealerWin() {
	//庄家收钱总额
	var val int64
	for seat, m := range t.DeskFree.Score3 {
		for _, v := range m {
			t.DeskFree.Score1[seat] += v
			val += v
		}
	}
	//庄家收钱转为正数
	if val < 0 {
		val *= -1
	}
	// if val > 0 {
	// 	//	抽成
	// 	val = t.drawcoin(t.DeskGame.Dealer, val)
	// }
	if val != 0 {
		t.DeskFree.Score1[uint32(pb.DESK_SEAT1)] = val
	}
	t.DeskFree.Carry += val //更新庄家收入携带
}

//.

// '闲家赔付
func (t *Desk) xianjiaJiesuan() {
	for seat, v := range t.DeskFree.Multiple {
		if v < 0 { //表示庄家输
			continue
		}
		tmp := t.getFreeBets(seat)
		t.DeskFree.Score3[seat] = make(map[string]int64)
		//赔付倍数
		var val int64
		if v > 1 {
			val = v - 1
		} else if v < -1 {
			val = (v * -1) - 1
		}
		if val != 0 { //表示庄家赢,且大于1倍从玩家身上扣赔付倍数
			for userid, betNum := range tmp {
				p := t.getPlayer(userid)
				if p == nil {
					glog.Errorf("userid %s not exist", userid)
					continue
				}
				coin := p.GetCoin()
				num := val * betNum
				if num > coin {
					num = coin
				}
				//扣除位置数
				t.sendCoin(userid, (-1 * int64(num)), int32(pb.LOG_TYPE6))
				t.DeskFree.Score3[seat][userid] = -1 * int64((betNum + num))
				t.DeskFree.Score2[userid] += -1 * int64((betNum + num))
			}
		} else {
			for userid, betNum := range tmp {
				t.DeskFree.Score3[seat][userid] = -1 * int64(betNum)
				t.DeskFree.Score2[userid] += -1 * int64(betNum)
			}
		}
	}
}

//.

// '庄家赔付
func (t *Desk) dealerJiesuan() {
	var num int64 //庄家赔付金额
	for seat, v := range t.DeskFree.Multiple {
		if v > 0 { //表示庄家赢
			continue
		}
		num += t.DeskFree.SeatBets[seat] * v * -1
	}
	if t.DeskFree.Carry >= num { //足够赔付
		t.dealerJiesuan1()
	} else { //不足赔付
		t.dealerJiesuan2(num)
	}
}

//.

// '足够赔付
func (t *Desk) dealerJiesuan1() {
	for seat, v := range t.DeskFree.Multiple {
		if v > 0 { //表示庄家赢
			continue
		}
		tmp := t.getFreeBets(seat)
		t.DeskFree.Score3[seat] = make(map[string]int64)
		var val int64 = v
		if v < 0 { //表示庄家输
			val = v * -1
		}
		for userid, betNum := range tmp {
			num := val * betNum
			if num > t.DeskFree.Carry {
				num = t.DeskFree.Carry
			}
			t.DeskFree.Carry -= num
			//	抽成, 赢利中抽取
			// num2 := t.drawcoin(userid, num)
			num2 := int64(0)
			val2 := int64(num2 + betNum)
			if val2 < 0 {
				val2 = 0
			}
			//扣除位置数
			t.sendCoin(userid, val2, int32(pb.LOG_TYPE6))
			t.DeskFree.Score3[seat][userid] = val2
			t.DeskFree.Score1[seat] += val2
			t.DeskFree.Score2[userid] += val2
			if num != 0 {
				t.DeskFree.Score1[uint32(pb.DESK_SEAT1)] -= int64(num)
			}
		}
	}
}

//.

// '不足赔付
func (t *Desk) dealerJiesuan2(num int64) {
	m := make(map[uint32]int64)
	for seat, v := range t.DeskFree.Multiple {
		if v > 0 { //表示庄家赢
			continue
		}
		//当前位置的总金额
		num1 := t.DeskFree.SeatBets[seat] * v * -1
		num2 := (num1 / num) * t.DeskFree.Carry
		m[seat] = num2 //位置分到金额
	}
	for seat, val := range m {
		tmp := t.getFreeBets(seat)
		betsNum := t.DeskFree.SeatBets[seat]
		t.DeskFree.Score3[seat] = make(map[string]int64)
		for userid, betNum := range tmp {
			num2 := (betNum / betsNum) * val //分到金额
			//	抽成, 赢利中抽取
			// num3 := t.drawcoin(userid, num2)
			num3 := int64(0)
			val2 := num3 + betNum //加上下注额
			if val2 < 0 {
				val2 = 0
			}
			//扣除位置数
			t.sendCoin(userid, val2, int32(pb.LOG_TYPE6))
			t.DeskFree.Score3[seat][userid] = val2
			t.DeskFree.Score1[seat] += val2
			t.DeskFree.Score2[userid] += val2
			if num2 != 0 {
				t.DeskFree.Score1[uint32(pb.DESK_SEAT1)] -= int64(num2)
			}
		}
	}
}

//.

// '获取对应位置下注列表
func (t *Desk) getFreeBets(seat uint32) map[string]int64 {
	return t.SeatRoleBets[seat]
}

//.

//.

// ' 游戏结束消息
func (t *Desk) resOverFree() *pb.ABFreeGameoverNtf {
	msg := &pb.ABFreeGameoverNtf{
		State:      t.state,
		Winner:     int32(t.ABDeskFree.Winner),
		SideWinner: int32(t.ABDeskFree.SideWinner),
		Userinfo:   t.freeSeatBetsMsg(),
	}
	for k, v := range t.ABScoreMap {
		bean := &pb.ABRoomScore{
			Userid: k,
			Seat:   t.getSeat(k),
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
		// 罐子流水
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
		if t.timer >= ReadyTime {
			t.timer = 0
			t.state = int32(pb.STATE_BET) // 发牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_BET)
			msg.During = t.Game.AB.BetTime
			t.selfPid.Tell(msg)
			return
		}
	case int32(pb.STATE_BET):
		// 下注状态
		if t.timer >= int(t.Game.AB.BetTime*2) {
			t.timer = 0
			// t.state = int32(pb.STATE_LEAD) // 开牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_LEAD)
			msg.During = 1
			t.selfPid.Tell(msg)
			t.tickStop = true
			return
		}
	case int32(pb.STATE_LEAD):
		// 开牌状态
		if t.timer >= t.DrawTime {
			t.timer = 0
			// t.state = int32(pb.STATE_OVER) // 结算状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_OVER)
			msg.During = FreeSettlementTime / 2
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
			msg.During = 2
			t.selfPid.Tell(msg)
			return
		}
	}
}

// 开牌
func (t *Desk) lead() {
	room, err := t.getRoom()
	if err != nil {
		// 异常牌局
		return
	}
	t.Game = room

	defer t.SetWinner()

	// 策略
	times := 0
	trigger := false
	t.selectStrategy()
	if len(t.ABStrategys) > 0 {
		for _, s := range t.ABStrategys {
			switch s.Id {
			case ANQS:
				trigger, times = t.ANQSStrategy()
			case ARTY:
				trigger, times = t.ARTYStrategy()
			}
			if trigger {
				t.ABDetail.StrategyId = s.Id
				break
			}
		}
	}

	if !trigger || times == 0 {
		// 没有触发策略，走调整系数
		// 调整系数，提前发牌
		t.MustLoseSeat()
		// 开始翻牌
		t.StartLead()
	} else {
		t.dispensePoker(times)
	}

	// 通知玩家
	ntf := new(pb.ABFreeLeadNtf)
	ntf.Andar = t.ABDeskFree.ACards
	ntf.Bahar = t.ABDeskFree.BCards
	defer t.broadcast(ntf)
}

// 根据下注总额获取对应的库存
func (t *Desk) getRoom() (data.Game, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := config.GetGames()
	var roomId int32 = 0
	for _, g := range games {
		if g.Gtype != int32(pb.ABAR) || g.Dtype != t.Dtype {
			continue
		}
		switch t.Dtype {
		case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
			roomId = g.AB.ControlRoom[0]
		// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
		// 	roomId = g.AB.NoviceRoom[0]
		case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
			int32(pb.DESK_TYPE_NEWBIEW): // 正常
			intervals := g.AB.BetInterval
			for i, v := range intervals {
				if pool >= int64(v[0]) && v[1] == -1 || (pool >= int64(v[0]) && pool < int64(v[1])) {
					roomId = g.AB.BetRoom[i][0]
					break
				}
			}
		}
	}
	room := config.GetGame(utils.String(roomId))
	if room.Id == "" {
		return room, fmt.Errorf("no find room, id:%d", roomId)
	}

	userid := t.GetOnlyOnePlayer()
	if userid != "" {
		role := t.roles[userid]
		if role.RegistArea == 0 {
			// A类玩家新手配置
			room.AB.NewbiewMode = room.AB.ANewbiewMode
		}
	}
	return room, nil
}

// 计算哪边赢或者开和
/*func  (t *Desk) calWinner(room data.Game) (winner, sideWinner uint32, err error) {
	// 配置结果
	// if env == "dev" && t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 	w := t.cfgWinner()
	// 	if w >= 0 {
	// 		if w == 0 {
	// 			t.freeDeal(true)
	// 		}
	// 		return uint32(w), nil
	// 	}
	// }
	factor, stock := t.getFinalFactor()
	if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
		// 点控
		factor, stock = t.getControlFactor()
	}
	t.DeskGame.FinaFactor, t.DeskGame.NowStock = factor, stock
	t.ABDetail.FinalFactor = factor

	// 选择策略
	strategy := t.selectStrategy()

	userid := t.GetOnlyOnePlayer()
	if b, ok := t.Bets[userid]; ok && b > 0 {
		role := t.roles[userid]
		if !role.PCSwitch && strategy == 0 && !utils.RandWan(t.Game.LOTTERY.UnRandomRate) {
			// 没点控，没策略，有概率随机开
			return 0, 0, errors.New("use nature")
		}
	}

	// B类玩家策略
	if w, sw, ok := t.WinStrategy_B(); ok {
		return w, sw, nil
	}

	var cardTypes []uint32
	switch strategy {
	case SF:
		// 劫富
		cardTypes = t.SFStrategyDraw()
	case JP:
		//济贫
		cardTypes = t.JPStrategyDraw()
		if len(cardTypes) > 0 {
			break
		}
		fallthrough
	default:
		cardTypes = t.drawCardType()
	}

	t.LDetail.StrategyId = strategy
	t.LDetail.CanWinScore = t.getMaxWinScore()
	// 最终系数
	// factor := t.getFinalFactor()
	// if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
	// 	// 点控
	// 	factor = t.getControlFactor()
	// }
	// t.DeskGame.FinaFactor = factor

	// 选择牌型
	// cardType := t.Game.AB.CardTypeID[len(t.Game.AB.FinalCoefficient)]
	// for i, v := range t.Game.AB.FinalCoefficient {
	// 	if factor <= float64(v)/100 {
	// 		cardType = t.Game.AB.CardTypeID[i]
	// 		break
	// 	}
	// }
	// // 新手房间走独立的胜率
	// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 	// 新手房间牌型另外计算
	// 	cardType = t.getNewBiewCardType()
	// }
	glog.Infof("andarbahar factor:%.2f, cardtype:%d,rid:%s,gameId:%s", factor, cardType, t.Rid, t.GameId)

	winner, sideWinner = t.winner(cardType)
	t.ABDeskFree.CardType = int(cardType)
	t.ABDeskFree.Winner, t.ABDeskFree.SideWinner = winner, sideWinner
	glog.Infof("andarbahar winner seat %d,,sideSeat:%d, gameid:%s", winner, sideWinner, t.GameId)

	return
}
*/
// 分配牌
func (t *Desk) dispensePoker(times int) {
	// 确认第几次开出KEY牌
	// times := algo.ABDispaterTimes(winner, sideWinner)

	if env == "dev" {
		// w := t.cfgWinner()
		// if w >= 0 {
		// 	times = w
		// }
	}

	// 先随机一张key牌
	// key := algo.ABGenKey(t.ABDeskFree.Joker)
	for i := 0; i < times; i++ {
		// 第50次就不拿了
		if i == 49 {
			break
		}

		v := t.DeskGame.Cards[i]
		if algo.ABKeyPoker(t.ABDeskFree.Joker, v) && i != times-1 {
			// 这是Key牌,但不是指定的次数
			// key = v
			// 换一张牌
			v = t.exchangePoker(i, v, false)
		} else if !algo.ABKeyPoker(t.ABDeskFree.Joker, v) && i == times-1 {
			// 最后一次,并且不是key牌,需要随机一张
			v = algo.ABGenKey(t.ABDeskFree.Joker)
		}

		// 最后一次必定是key牌
		if i == 48 {
			v = algo.ABGenKey(t.ABDeskFree.Joker)
		}

		if i%2 == 0 {
			// andar
			t.ACards = append(t.ACards, v)
		} else {
			// bahar
			t.BCards = append(t.BCards, v)
		}
	}

	// 计算开牌时间q
	t.ABDeskFree.DrawTime = (len(t.ACards) + len(t.BCards)) + 2
}

// 开牌结果msg
func (t *Desk) LeadMsg() *pb.ABFreeRoomOver {
	msg := new(pb.ABFreeRoomOver)
	msg.ValueA = t.ABDeskFree.ACards
	msg.ValueB = t.ABDeskFree.BCards
	msg.Joker = t.ABDeskFree.Joker
	return msg
}

// 玩家下注总额
func (t *Desk) userSeatBetMsg(userid string) []*pb.ABFreeUserBet {
	bets := make([]*pb.ABFreeUserBet, 0)
	for k, v := range t.DeskFree.ABSeatRoleBets {
		value := v[userid]
		msg := new(pb.ABFreeUserBet)
		msg.Seat = int32(k)
		msg.Bets = uint32(value.GetSum())
		bets = append(bets, msg)
	}
	return bets
}

// 下注总额
func (t *Desk) seatBetMsg() []*pb.ABFreeUserBet {
	bets := make([]*pb.ABFreeUserBet, 0)
	for k, v := range t.DeskFree.ABSeatRoleBets {
		for _, value := range v {
			msg := new(pb.ABFreeUserBet)
			msg.Seat = int32(k)
			msg.Bets = uint32(value.GetSum())
			bets = append(bets, msg)
		}
	}
	return bets
}

// 更新排行榜
func (t *Desk) updateRank(userid string) {
	rank := t.freeRankMsg(20)
	if userid != "" {
		update := false
		for i, lu := range rank {
			if i > 6 {
				break
			}
			if lu.Userid == userid {
				update = true
				rank = append(rank[:i], rank[i+1:]...)
			}
		}
		if update {
			msg := new(pb.ABFreeRankNtf)
			msg.Rank = rank
			t.broadcast4(msg)
		}
		return
	}
	msg := new(pb.ABFreeRankNtf)
	msg.Rank = rank
	t.broadcast4(msg)
}

// 事件
func (t *Desk) eventPost(userid string, eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("andarbahar event Marshal fail")
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
	// 开joker牌
	t.choiceJoker()
}

// 对局详情
func (t *Desk) gameDetail(details []*data.ABUserDetail) {
	if len(details) <= 0 {
		// 没人的不记
		return
	}
	// de := data.ABDetail{
	// 	Winner:       t.ABDeskFree.Winner,
	// 	SideWinner:   t.ABDeskFree.SideWinner,
	// 	Joker:        t.ABDeskFree.Joker,
	// 	ACards:       t.ACards,
	// 	BCards:       t.BCards,
	// 	JackpotValue: t.getJackpotValue(),
	// 	Bets:         t.CashBets + t.CoinBets,
	// 	UserDetail:   details,
	// }
	t.ABDetail.Winner = t.ABDeskFree.Winner
	t.ABDetail.SideWinner = t.ABDeskFree.SideWinner
	t.ABDetail.Joker = t.ABDeskFree.Joker
	t.ABDetail.ACards = t.ACards
	t.ABDetail.BCards = t.BCards
	t.ABDetail.JackpotValue = t.getJackpotValue()
	t.ABDetail.Bets = t.CashBets + t.CoinBets
	t.ABDetail.UserDetail = details

	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	t.ABDetail.PlayerWin = playerWin
	// 保存详情
	t.saveDetail(t.ABDetail)
}

// 保存详情
func (t *Desk) saveDetail(de data.ABDetail) {
	detail := data.Detail{
		WaterId:      t.GameId,
		BeginTime:    t.BeginTime,
		EndTime:      utils.BsonNow().Unix(),
		Gtype:        int32(pb.ABAR),
		Rtype:        t.Rtype,
		Gmode:        t.Gmode,
		RoomId:       t.Rid,
		DeskId:       t.Game.Id,
		CardTypeId:   t.ABDeskFree.CardType,
		PlayerFactor: int(t.DeskGame.FinaFactor * 100),
		ABDetail:     &de,
	}
	players := make([]string, 0)
	for _, ld := range de.UserDetail {
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

// func (t *Desk) maxOutCash(out int64) bool {
// 	if t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
// 		return false
// 	}
// 	bean := table.GetTables().NewbieTable.Get()
// 	for _, r := range t.roles {
// 		if r.OutDiamond+out <= int64(bean.OutCashLimited) {
// 			return true
// 		}
// 	}
// 	return false
// }

// 新手人机结算
func (t *Desk) robotSettlement(id string, seat uint32, c data.Currency) bool {
	if user, ok := t.NewbiewRobot[id]; ok {
		if seat == t.ABDeskFree.Winner {
			odds := t.ABDeskFree.OddsMap[t.ABDeskFree.Winner] // 赔率
			user.AddDiamond(c.Diamond * int64(odds) / 100)
			t.ABScoreMap[id] = c.Scale(float64(odds) / 100)
		} else if seat == t.ABDeskFree.SideWinner {
			odds := t.ABDeskFree.OddsMap[t.ABDeskFree.SideWinner] // 赔率
			user.AddDiamond(c.Diamond * int64(odds) / 100)
			t.ABScoreMap[id] = c.Scale(float64(odds) / 100)
		} else {
			// 记录输赢分
			t.ABScoreMap[id] = c.Scale(-1)
		}
		return true
	}
	return false
}

// 历史记录
func (t *Desk) buildHistory() *pb.ABHistory {
	history := &pb.ABHistory{
		Winner: t.ABWinnerSeat,
	}

	// joker牌对应的位置
	seats := t.ABWinnerSeat[len(t.ABWinnerSeat)-len(t.ABJoker):]
	if len(seats) <= 0 {
		return history
	}

	for i, joker := range t.ABJoker {
		bean := &pb.ABJokerHis{
			Winner: seats[i],
			Joker:  joker,
		}
		history.Joker = append(history.Joker, bean)
	}
	return history
}

// 中奖的牌值
func (t *Desk) getJackpotValue() uint32 {
	if t.ABDeskFree.Winner == Andar {
		return t.ACards[len(t.ACards)-1]
	}
	return t.BCards[len(t.BCards)-1]
}

// 把key牌换为非key牌
func (t *Desk) exchangePoker(index int, poker uint32, key bool) uint32 {
	if algo.ABKeyPoker(t.ABDeskFree.Joker, poker) {
		if key {
			// key牌
			return poker
		}
		// 非key牌
		t.DeskGame.Cards = append(t.DeskGame.Cards[:index], t.DeskGame.Cards[index+1:]...)
		poker = t.DeskGame.Cards[index]
		t.DeskGame.Cards = append(t.DeskGame.Cards, poker)
		return t.exchangePoker(index+1, poker, key)
	}
	return poker
}

func (t *Desk) WinStrategy_B() (uint32, uint32, bool) {
	if t.Dtype != int32(pb.DESK_TYPE_NORMAL_B) && t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
		return 0, 0, false
	}

	for _, role := range t.roles {
		if role.Robot {
			continue
		}
		bet := t.Bets[role.Userid]
		factor := t.UserFactorMap[role.Userid]
		if handler.FreeWinStrategy_B(role.User, bet, factor) {
			if bet < 1000 {
				return 0, 0, false
			}
			for k, v := range t.ABSeatRoleBets {
				if 3 <= k && k <= 7 {
					b := v[role.Userid].GetSum()
					if b > 0 && b != bet {
						return 0, 0, false
					}
					if b == 0 {
						continue
					}
					// 玩家只要在3-7位置下注超过10就必赢
					t.send2userid(role.Userid, &pb.TriggerFreeWelfare{Userid: role.Userid})
					return uint32(utils.RandInt32N(2) + 1), k, true
				}
			}
		}
	}
	return 0, 0, false
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

// vim: set foldmethod=marker foldmarker=//',//.:
