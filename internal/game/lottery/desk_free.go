package lottery

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
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
			glog.Infof("init CP id:%s", t.DeskData.Game.Id)
			bean := config.GetGame(t.DeskData.Game.Id)
			if bean.Id != "" {
				t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
				// glog.Infof("init lottery data:%#v", games)
				break
			}
		}
	}

	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//下注时间
	t.BetTime = int(t.Game.LOTTERY.BetTime)
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
	t.BackOutDiamond = make(map[string]int64)

	t.DeskFree.LotteryDesk = &data.LotteryDesk{
		//位置下注详细
		LSeatRoleBets: make(map[uint32]map[string]data.Currency),
		LScoreMap:     make(map[string]data.Currency),
		CPRobotSeat:   make(map[string]uint32),
		CPRobotBets:   make(map[string]int64),
	}
	//玩家个人系数
	t.DeskFree.UserFactorMap = make(map[string]float64)
	t.state = int32(pb.STATE_READY)
	//初始化赔率
	t.initOdds()
}

// 初始化新手模式
func (t *Desk) newbiewInit() {
	// betweights := t.Game.LOTTERY.Robot.BetWeight
	// choices := make([]utils.Choice, 0)
	// for i, v := range betweights {
	// 	if i == 5 {
	// 		continue
	// 	}
	// 	choices = append(choices, utils.Choice{Weight: int(v), Item: uint32(i + 1)})
	// }
	room := t.DeskData.Game
	initScore := room.LOTTERY.Robot.InitScore
	t.NewbiewRobot = make(map[string]*data.User)
	// 奖池
	t.CPJackpot = utils.RandInt64N(300000000) + 50000000
	// 新手房间,生成假人,20个
	// vip := config.GetVipRobot()[0]
	for i := 0; i < 20; i++ {
		user := &data.User{
			Nickname:   login.RandName(),
			Userid:     fmt.Sprintf("%d", i+1),
			Photo:      strconv.Itoa(utils.RandIntN(30) + 1),
			Coin:       int64(utils.RandMN(int(initScore[0]), int(initScore[1]))),
			FreeWinMap: make(map[int32][]data.FreeWin),
			Robot:      true,
		}
		t.NewbiewRobot[user.Userid] = user
	}
	// 50条历史记录
	for i := 0; i < 50; i++ {
		t.shuffle()
		ctype := algo.HuaType(t.DeskGame.Cards[:3])
		// c, _ := utils.WeightedChoice(choices)
		t.CPHistory = append(t.CPHistory, ctype)
	}
	// bigwinner
	zeroTime := (utils.TimestampToday(location) - 3600*24) * 1000
	now := utils.LocalTime().UnixMilli()
	ctimes := make([]int64, 0)
	for i := 0; i < 10; i++ {
		t.shuffle()
		cards, _ := algo.GetCard(algo.BaoZi, t.DeskGame.Cards)
		draw := (utils.RandInt64N(300000000) + 50000000) * 2000 / 10000
		bean := data.CPJackpot{
			Nickname: login.RandName(),
			Bet:      (utils.RandInt64N(20) + 1) * 1000,
			Get:      (utils.RandInt64N(1000) + 1000) * draw / 10000,
			Draw:     draw,
			Cards:    cards,
			People:   (utils.RandIntN(11)+10)*utils.RandIntN(4) + 2,
			// Ctime:    utils.RandInt64N(now-zeroTime) + zeroTime,
		}
		ctimes = append(ctimes, utils.RandInt64N(now-zeroTime)+zeroTime)
		head := handler.GetHead()
		if head != nil {
			bean.Photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/man/%d.jpg", head.Id)
			bean.Nickname = head.Name
			// 归还
			handler.ReturnHead(head.Sex, head.Id)
		}
		t.CPBigWinners = append(t.CPBigWinners, bean)
	}
	sort.Slice(t.CPBigWinners, func(i, j int) bool {
		return t.CPBigWinners[i].Draw > t.CPBigWinners[j].Draw
	})
	sort.Slice(ctimes, func(i, j int) bool {
		return ctimes[i] < ctimes[j]
	})

	for i, ctime := range ctimes {
		t.CPBigWinners[i].Ctime = ctime
	}
	// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) || t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
	// }
}

// '进入房间响应消息
func (t *Desk) freeEnterMsg(userid string) *pb.LotteryEnterRsp {
	msg := new(pb.LotteryEnterRsp)
	//房间数据
	msg.Roominfo = handler.PackCPFreeRoom(t.DeskData)
	t.freeRoomDataMsg(msg.Roominfo)
	//玩家下注的位置
	msg.UserBets = t.userBetsAllSeat(userid)
	//大赢家
	msg.Winners = t.bigWinnerList()
	//奖池
	msg.Jackpot = t.CPJackpot
	return msg
}

func (t *Desk) freeRoomDataMsg(msg *pb.LotteryRoom) {
	msg.State = t.state
	// 输赢记录20条
	msg.Winer = t.DeskFree.CPHistory
	// 位置对应下注
	for seat, v := range t.SeatBets {
		bean := &pb.LotterySeatBets{
			Seat: seat,
			Bets: v,
		}
		msg.Bets = append(msg.Bets, bean)
	}

	tt := 0
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		switch t.state {
		case int32(pb.STATE_READY):
			tt = 1 - t.timer
		case int32(pb.STATE_BET):
			tt = t.BetTime - t.timer
		case int32(pb.STATE_LEAD):
			tt = FreeLeadTime - t.timer
			// 开牌阶段要把开的牌发下去
			msg.Cards = t.LCards
			msg.CardType = algo.HuaType(t.LCards)
		case int32(pb.STATE_OVER):
			tt = FreeSettlementTime - t.timer
			// 结算阶段要把开的牌发下去
			msg.Cards = t.LCards
			msg.CardType = algo.HuaType(t.LCards)
		default:
			tt = t.BetTime - t.timer
		}
	}
	if tt < 0 {
		tt = 0
	}
	msg.Timer = uint32(tt)
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
	if num <= 0 {
		return pb.OperateError
	}
	if num > role.Diamond {
		return pb.NotEnoughCoin
	}

	if t.LSeatRoleBets[seatBet] != nil {
		c := t.LSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		if limit > int64(t.DeskData.Game.LOTTERY.BetLimit[seatBet-1]) {
			return pb.BetTopLimit //下注限制
		}
	}

	var cashBet, coinBet int64 = num, 0
	// cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	// coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, 0, (-1 * cashBet), int32(pb.LOG_TYPE90), fmt.Sprintf("彩票%s房间下注", t.DeskData.Rid))
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
	if m, ok := t.LSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Diamond += cashBet
		c.Coin += coinBet
		m[userid] = c
		t.LSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Diamond: cashBet,
			Coin:    coinBet,
		}
		t.LSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	role.NoBetTimes = 0 // 不下注次数清零

	// 奖池增加
	var addJackpot int64 = 0
	if user.Robot {
		addJackpot = num * 500 / 10000
		t.CPJackpot += addJackpot
		// 最大值为500w
		t.CPJackpot = int64(math.Min(500000000, float64(t.CPJackpot)))
	}

	msg, ntf := resFreeBet(seatBet, cashBet+coinBet,
		t.DeskFree.SeatBets[seatBet], betsNum, addJackpot, userid)
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
	if t.LSeatRoleBets[seatBet] != nil {
		c := t.LSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		if limit > int64(t.DeskData.Game.LOTTERY.BetLimit[seatBet-1]) {
			return pb.BetTopLimit //下注限制
		}
	}
	user.AddCoin(-num)
	t.DeskFree.Bets[userid] += num      //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num //当前位置总下注额(没扣税的)
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.LSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Coin += num
		m[userid] = c
		t.LSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Coin: num,
		}
		t.LSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}

	// 奖池增加
	var addJackpot int64 = 0
	if user.Robot {
		addJackpot = num * 500 / 10000
		t.CPJackpot += addJackpot
		// 最大值为500w
		t.CPJackpot = int64(math.Min(500000000, float64(t.CPJackpot)))
	}
	_, ntf := resFreeBet(seatBet, num,
		t.DeskFree.SeatBets[seatBet], betsNum, addJackpot, userid)
	t.broadcast(ntf)
	return pb.OK
}

// 下注消息
func resFreeBet(beseat uint32, val, coin,
	bets, addJackpot int64, userid string) (*pb.LotteryBetRsp, *pb.LotteryBetNtf) {
	rsp := &pb.LotteryBetRsp{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
	}
	ntf := &pb.LotteryBetNtf{
		Beseat:     beseat,
		Value:      uint32(val),
		Coin:       coin,
		AddJackpot: int32(addJackpot),
		Userid:     userid,
	}
	return rsp, ntf
}

// '洗牌
func (t *Desk) shuffle() {
	rand.Seed(time.Now().UnixNano())
	d := make([]uint32, algo.NumCard, algo.NumCard)
	copy(d, algo.NiuCARDS)
	//测试暂时去掉洗牌
	for i := range d {
		j := rand.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	t.DeskGame.Cards = d
}

// '结束游戏
func (t *Desk) freeGameOver() {
	//赔付
	details, output := t.gameSettlement()
	//对局详情
	t.gameDetail(details, output)
	//打印信息
	t.printOver()
	//结束消息
	t.resOverFree()
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
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "70033")
	}
}

//.

// .结账
func (t *Desk) gameSettlement() ([]data.CPUserDetail, int64) {
	details := make([]data.CPUserDetail, 0)
	scoreMap := make(map[string]data.Currency)
	var cashStock, coinStock, mcashTaxStock, acashTaxStock, mcoinTaxStock, acoinTaxStock int64 = 0, 0, 0, 0, 0, 0 // 库存变化

	winner := t.getWinner()
	roleBets := t.LSeatRoleBets
	odds := float64(t.DeskFree.Multiple[winner]) / 100 // 赔率

	//豹子开奖每人可分得的奖励
	pool := t.drawJackpot(winner)
	bzMulti := utils.RandInt32N(51) + 300

	for k, roles := range roleBets {
		// 赢的玩家发奖
		for id, v := range roles {
			role := t.roles[id]
			if role == nil {
				if !t.robotSettlement(id, k, winner, v, pool, bzMulti) {
					// 新手状态人机结算
					glog.Errorf("lottery Settlement role is nil, id:%s", id)
				}
				continue
			}
			winBets := v.Scale(-1)
			if k == winner {
				drawReward := t.sharePrize(winner, pool, id, bzMulti)
				w := v.Scale(odds)
				if role.Robot {
					// 人机加奖励金
					w.Coin += drawReward
				} else {
					// 玩家加彩金
					w.Diamond += drawReward
				}
				winBets.Merge(w)
				// 返奖
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE91), fmt.Sprintf("彩票房间%s赢分", t.Game.Id))
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
		var mcash, mcoin, acash, acoin, realCash int64 = 0, 0, 0, 0, c.Diamond
		if c.GetSum() > 0 {
			// 赢了钱,扣税
			mcash, mcoin = t.getTax(realCash, c.Coin, true)
			// 赢了钱,扣税
			acash, acoin = t.getTax(realCash-mcash, c.Coin-mcoin, false)
			// 扣明税
			// t.sendCurrency(k, -mcoin, -mcash, int32(pb.LOG_TYPE92), fmt.Sprintf("彩票房间%s扣税", t.Game.Id))
		} else {
			// 扣暗税
			realCash = -int64(math.Max(float64(-realCash-role.GiveDiamond), 0))
			acash, acoin = t.getTax(realCash, c.Coin, false)
		}
		// acoin := t.getTax(c.Coin-mcoin, false)
		// 玩家输赢记录
		t.lotteryRecord(k, c.GetSum())
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
				Gtype: uint32(pb.LOTTERY),
				Win:   c.GetSum() > 0,
			}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType: int32(pb.LOTTERY),
				Win:      c.GetSum() > 0,
			})
			// 对局详情
			details = append(details, t.createDetail(k, c, 0, 0, acash, acoin)) //, mcash, mcoin, acash, acoin))
		}
		// 记录输赢分
		// c.Merge(data.Currency{Diamond: -mcash, Coin: -mcoin})
		t.LScoreMap[k] = c

		if !role.Robot {
			t.strategySettlement(k)
		}
	}
	// 观察着
	for _, v := range t.CPObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			details = append(details, t.createDetail(v, data.Currency{}, 0, 0, 0, 0))
		}
	}
	// 游戏开奖记录
	t.addHistory(winner)
	// 大赢家记录
	t.bigWinnerRecord(winner, pool, bzMulti)
	// 库存变化
	t.changeStock(-cashStock, -coinStock, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock, t.Game.Id)
	return details, pool
}

// 输赢记录
func (t *Desk) lotteryRecord(useid string, num int64) {
	bets := t.DeskFree.Bets[useid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.LOTTERY), Score: num, Bet: bets}
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
	if w, ok := role.FreeWinMap[int32(pb.LOTTERY)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.LOTTERY)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.LOTTERY)]++

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    role.Pid,
			Userid:     role.Userid,
			Gtype:      int32(pb.LOTTERY),
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
			glog.Error("publish user lottery bets error", err)
		}

		log := handler.GameFlowWaterLog(bets, int(pb.LOTTERY), 0, len(role.RechargeTarge), useid)
		myactor.Logger().Tell(log)
	}

}

// 开奖记录
func (t *Desk) addHistory(win uint32) {
	t.CPHistory = append(t.CPHistory, win)
	size := len(t.CPHistory)
	if size > 50 { // 只保留50条
		t.CPHistory = t.CPHistory[size-50:]
	}
}

// 大赢家记录
func (t *Desk) bigWinnerRecord(winner uint32, pool int64, multi int32) {
	if winner != algo.BaoZi || pool == 0 {
		return
	}
	userid := ""
	var winscore int64 = 0
	betmap := t.LSeatRoleBets[algo.BaoZi]
	if betmap == nil || len(betmap) <= 0 {
		return
	}
	for k := range betmap {
		win := t.LScoreMap[k]
		if win.GetSum() > winscore {
			userid = k
			winscore = win.GetSum()
		}
	}
	if userid == "" {
		return
	}
	role := new(data.User)
	if r, ok := t.roles[userid]; ok {
		role = r.User
	}
	if r, ok := t.NewbiewRobot[userid]; ok {
		role = r
	}
	if role == nil || role.Userid == "" {
		return
	}
	bean := data.CPJackpot{
		GameId:   t.GameId,
		Userid:   userid,
		Nickname: role.Nickname,
		Photo:    role.Photo,
		Bet:      betmap[userid].GetSum(),
		Get:      winscore,
		Draw:     pool,
		Ctime:    utils.BsonNow().UnixMilli(),
		Cards:    t.LCards,
		People:   len(betmap) * (utils.RandIntN(4) + 2),
	}
	head := handler.GetHead()
	if head != nil {
		bean.Photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/man/%d.jpg", head.Id)
		bean.Nickname = head.Name
		// 归还
		handler.ReturnHead(head.Sex, head.Id)
	}
	t.CPBigWinners = append(t.CPBigWinners, bean)
	size := len(t.CPBigWinners)
	if len(t.CPBigWinners) > 10 {
		t.CPBigWinners = t.CPBigWinners[size-10:]
	}
}

// . 对局详情
func (t *Desk) createDetail(userid string, score data.Currency, mcash, mcoin, acash, acoin int64) data.CPUserDetail {
	bet := t.Bets[userid]
	detail := data.CPUserDetail{
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
	detail.BeforeScore = role.GetScore() - score.GetSum() + mcash
	detail.BeforeCash = role.Diamond - score.Diamond + mcash
	detail.AfterScore = role.GetScore()
	detail.AfterCash = role.Diamond
	// if score.GetSum() >= 0 {
	// 	detail.AfterScore = role.GetScore() + bet + score.GetSum() - mcash
	// 	detail.AfterCash = role.Diamond + bet + score.Diamond - mcash
	// } else {
	// 	detail.AfterScore = role.GetScore()
	// 	detail.AfterCash = role.Diamond
	// }
	// detail.AfterBonus = role.Coin
	// detail.BeforeBonus = role.Coin - score.Coin
	// detail.BeforeCash = role.Diamond + bet
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

// '获取对应位置下注列表
func (t *Desk) getFreeBets(seat uint32) map[string]int64 {
	return t.SeatRoleBets[seat]
}

//.

// ' 游戏结束消息
func (t *Desk) resOverFree() {
	for k := range t.roles {
		msg := &pb.LotteryGameOverNtf{
			Jackpot:  t.CPJackpot,
			CardType: algo.HuaType(t.LCards),
			Winner:   t.gameBigWinner(),
		}
		msg.Score = t.LScoreMap[k].GetSum()
		t.send2userid(k, msg)
	}
	for k, v := range t.LScoreMap {
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
			t.state = int32(pb.STATE_BET) // 下注状态
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
			msg.During = 1
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
	room, err := t.getRoom()
	if err != nil {
		// 异常牌局
		return
	}
	t.Game = room
	// 洗牌
	t.shuffle()
	// 发牌
	t.calWinner()
	// 分牌
	// t.dispensePoker(winner)

	// 通知玩家
	ntf := new(pb.LotteryLeadNtf)
	ntf.Cards = t.LCards
	ntf.CardType = algo.HuaType(t.LCards)
	t.broadcast(ntf)
}

// 根据下注总额获取对应的库存
func (t *Desk) getRoom() (data.Game, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := config.GetGames()
	var roomId int32 = 0
	for _, g := range games {
		if g.Gtype != int32(pb.LOTTERY) || g.Dtype != t.Dtype {
			continue
		}
		switch t.Dtype {
		case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
			roomId = g.LOTTERY.ControlRoom[0]
		// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
		// 	roomId = g.LOTTERY.NoviceRoom[0]
		case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
			int32(pb.DESK_TYPE_NEWBIEW): // 正常
			intervals := g.LOTTERY.BetInterval
			for i, v := range intervals {
				if pool >= int64(v[0]) && v[1] == -1 || (pool >= int64(v[0]) && pool < int64(v[1])) {
					roomId = g.LOTTERY.BetRoom[i][0]
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
			room.LOTTERY.NewbiewMode = room.LOTTERY.ANewbiewMode
		}
	}
	return room, nil
}

// 计算哪边赢或者开和
func (t *Desk) calWinner() {
	// 策略
	trigger, random := false, false
	t.selectStrategy()
	if len(t.CPStrategys) > 0 {
		for _, s := range t.CPStrategys {
			switch s.Id {
			case LWJY:
				trigger, random = t.LWJYStrategy()
			case LYQN:
				trigger = t.LYQNStrategy()
			}
			if trigger {
				t.LDetail.StrategyId = s.Id
				t.CPTriggerStrategy = s.Id
				break
			}
		}
	}
	if !trigger || random {
		// 策略没生效
		// loseSeat := t.MustLoseSeat()
		// 开始翻牌
		t.StartLead()
	}

	// 开奖权重
	// drawPro := t.Game.LOTTERY.DrawPro
	// 配置结果
	// if env == "dev" {
	// 	w := t.cfgWinner()
	// 	if w > 0 {
	// 		return uint32(w), nil
	// 	}
	// }

	// 最终系数
	// factor, stock := t.getFinalFactor()
	// if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
	// 	// 点控
	// 	factor, stock = t.getControlFactor()
	// }
	// t.DeskGame.FinaFactor, t.DeskGame.NowStock = factor, stock
	// t.LDetail.FinalFactor = factor
	// glog.Infof("cp FinalFactor %f, stock:%d, rid:%s, gameid:%s", factor, stock, t.Rid, t.GameId)

	/* if t.Dtype == int32(pb.DESK_TYPE_NORMAL) || t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		// A类和B类的正常玩家自然概率
		return t.NatureWinner(), nil
	}
	// B类策略

	if w, ok := t.WinStrategy_B(); ok {
		return w, nil
	} */

	// 平台胜率(万分比)
	/* winPro := room.LOTTERY.WinningPro[len(room.LOTTERY.FinalCoefficient)]
	for i, v := range room.LOTTERY.FinalCoefficient {
		if factor <= float64(v)/100 {
			winPro = room.LOTTERY.WinningPro[i]
			break
		}
	}
	// 新手房间走独立的胜率
	if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
		winPro = 10000 - t.getNewBiewWinning()
	}

	// 哪方赢
	var cardTypes []uint32
	if utils.RandWan(winPro) {
		glog.Infof("lottery player lose, winPro %d stockid:%s, rid:%s, gameid:%s", winPro, t.Game.Id, t.Rid, t.GameId)
		// 平台赢
		// 计算所有平台赢的牌型
		cardTypes = t.drawCardType(true)
	} else {
		glog.Infof("lottery player win, winPro %d stockid:%s, rid:%s, gameid:%s", (10000 - winPro), t.Game.Id, t.Rid, t.GameId)
		// 玩家赢
		// 计算所有平台输的牌型
		cardTypes = t.drawCardType(false)
	} */

	// 奖池超额
	// if t.jackpotExceed() {
	// 	glog.Infof("jackpotExceed gameid:%s", t.GameId)
	// 	return algo.BaoZi, nil
	// }

	// // 策略
	// strategy := t.selectStrategy()

	// userid := t.GetOnlyOnePlayer()
	// if b, ok := t.Bets[userid]; ok && b > 0 {
	// 	role := t.roles[userid]
	// 	if !role.PCSwitch && strategy == 0 && !utils.RandWan(t.Game.LOTTERY.UnRandomRate) {
	// 		// 没点控，没策略，有概率随机开
	// 		return t.NatureWinner(), nil
	// 	}
	// }

	// var cardTypes []uint32
	// switch strategy {
	// case SF:
	// 	// 劫富
	// 	cardTypes = t.SFStrategyDraw()
	// case JP:
	// 	//济贫
	// 	cardTypes = t.JPStrategyDraw()
	// 	if len(cardTypes) > 0 {
	// 		break
	// 	}
	// 	fallthrough
	// default:
	// 	cardTypes = t.drawCardType()
	// }

	// t.LDetail.StrategyId = strategy
	// t.LDetail.CanWinScore = t.getMaxWinScore()

	// // 计算开哪个位置
	// choices := make([]utils.Choice, 0)
	// for _, v := range cardTypes {
	// 	if int(v) > len(drawPro) {
	// 		continue
	// 	}
	// 	choices = append(choices, utils.Choice{Weight: drawPro[v-1], Item: v})
	// }

	// // 随机牌型
	// choice, err := utils.WeightedChoice(choices)
	// if err != nil {
	// 	// 没找到,随机一个
	// 	t.LDetail.IsRandom = true
	// 	glog.Error("cp no found cardtype,len:", len(cardTypes))
	// 	index, _ := utils.ChoiceIntIndex(drawPro)
	// 	return uint32(index + 1), nil
	// }

	// glog.Infof("lottery winner cardtype %d, gameid:%s, rid:%s", choice.Item, t.GameId, t.Rid)
	// return choice.Item.(uint32), nil
}

// 分配牌
func (t *Desk) dispensePoker(cardType uint32) {
	t.LCards, t.DeskGame.Cards = algo.GetCard(cardType, t.DeskGame.Cards)
}

// .赢家
func (t *Desk) getWinner() uint32 {
	return algo.HuaType(t.LCards)
}

// 事件
func (t *Desk) eventPost(userid string, eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("lottery event Marshal fail")
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
func (t *Desk) gameDetail(details []data.CPUserDetail, output int64) {
	if len(details) <= 0 {
		// 新手房没下注的不记
		return
	}
	t.LDetail.Winner = t.getWinner()
	t.LDetail.Cards = t.LCards
	t.LDetail.CardType = algo.HuaType(t.LCards)
	t.LDetail.Bets = t.CashBets + t.CoinBets
	t.LDetail.JackpotOutput = output
	t.LDetail.UserDetail = details
	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	t.LDetail.PlayerWin = playerWin
	// 保存详情
	t.saveDetail()
}

// 保存详情
func (t *Desk) saveDetail() {
	detail := data.Detail{
		WaterId:      t.GameId,
		BeginTime:    t.BeginTime,
		EndTime:      utils.BsonNow().Unix(),
		Gtype:        int32(pb.LOTTERY),
		RoomId:       t.Rid,
		DeskId:       t.Game.Id,
		PlayerFactor: int(t.DeskGame.FinaFactor * 100),
		CPDetail:     &t.LDetail,
	}
	players := make([]string, 0)
	for _, ld := range t.LDetail.UserDetail {
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

func (t *Desk) maxOutCash(out int64) bool {
	if t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
		return false
	}
	// bean := table.GetTables().NewbieTable.Get()
	for _, r := range t.roles {
		if r.OutDiamond+out <= int64(t.Game.LOTTERY.NewbiewMode.OutCashLimited) {
			return true
		}
	}
	return false
}

// 新手人机结算
func (t *Desk) robotSettlement(id string, seat, winner uint32, c data.Currency, pool int64, multi int32) bool {
	if user, ok := t.NewbiewRobot[id]; ok {
		score := t.LScoreMap[id]
		if seat == winner {
			drawReward := t.sharePrize(winner, pool, id, multi)
			odds := float64(t.DeskFree.Multiple[winner]) / 100 // 赔率
			user.AddCoin(c.GetSum()*int64(odds) + drawReward)
			score.Merge(c.Scale(float64(odds)))
			score.Coin += drawReward
			t.LScoreMap[id] = score
		} else {
			// 记录输赢分
			score.Merge(c.Scale(-1))
			t.LScoreMap[id] = score
		}
		return true
	}
	return false
}

// 获取玩家每个位置下注
func (t *Desk) userBetsAllSeat(userid string) []*pb.LotterySeatBets {
	bets := make([]*pb.LotterySeatBets, 0)
	for k, v := range t.LSeatRoleBets {
		if bet, ok := v[userid]; ok {
			bean := &pb.LotterySeatBets{Seat: k, Bets: bet.GetSum()}
			bets = append(bets, bean)
		}
	}
	return bets
}

// 大赢家列表
func (t *Desk) bigWinnerList() []*pb.LotteryBigWinner {
	beans := make([]*pb.LotteryBigWinner, 0)
	for _, w := range t.CPBigWinners {
		b := &pb.LotteryBigWinner{
			Userid:     w.Userid,
			Nickname:   w.Nickname,
			Photo:      w.Photo,
			Bet:        w.Bet,
			Get:        w.Get,
			Jackpot:    w.Draw,
			Ctime:      w.Ctime,
			Cards:      w.Cards,
			DrawPeople: int32(w.People),
		}
		beans = append(beans, b)
	}

	sort.Slice(beans, func(i, j int) bool {
		return beans[i].Ctime < beans[j].Ctime
	})

	return beans
}

// 获取玩家位置对应下注
func (t *Desk) getSeatBetsOfUserid(userid string) map[string]int64 {
	seatBetMap := make(map[string]int64)
	for k, bets := range t.LSeatRoleBets {
		if b, ok := bets[userid]; ok {
			seatBetMap[utils.String(k)] += b.GetSum()
		}
	}
	return seatBetMap
}

func (t *Desk) sharePrize(winner uint32, pool int64, userid string, multi int32) int64 {
	if pool <= 0 {
		return 0
	}
	allBet := t.SeatBets[winner]                        // 总下注
	userBet := t.LSeatRoleBets[winner][userid].GetSum() // 玩家下注
	if allBet*int64(multi) <= pool {
		return userBet * int64(multi)
	}
	// 赔付超过jackpot
	return int64(math.Floor(float64(userBet) / float64(allBet) * float64(pool)))
}

func (t *Desk) gameBigWinner() *pb.LotteryBigWinner {
	for _, c := range t.CPBigWinners {
		if c.GameId != t.GameId {
			continue
		}
		return &pb.LotteryBigWinner{
			Userid:     c.Userid,
			Nickname:   c.Nickname,
			Photo:      c.Photo,
			Get:        c.Get,
			Bet:        c.Bet,
			Ctime:      c.Ctime,
			Jackpot:    c.Draw,
			Cards:      c.Cards,
			DrawPeople: int32(c.People),
		}
	}
	return nil
}

func (t *Desk) WinStrategy_B() (uint32, bool) {
	if t.Dtype != int32(pb.DESK_TYPE_NORMAL_B) && t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
		return 0, false
	}

	for _, role := range t.roles {
		if role.Robot {
			continue
		}
		bet := t.Bets[role.Userid]
		factor := t.UserFactorMap[role.Userid]
		if handler.FreeWinStrategy_B(role.User, bet, factor) {
			if bet < 2000 {
				return 0, false
			}
			for k, v := range t.LSeatRoleBets {
				if k >= algo.GaoPai && k <= algo.TongHuaShun {
					b := v[role.Userid].GetSum()
					if b > 0 && b != bet {
						return 0, false
					}
					if b == 0 {
						continue
					}
					// 玩家只要下注超过20就开和
					t.send2userid(role.Userid, &pb.TriggerFreeWelfare{Userid: role.Userid})
					return k, true
				}
			}
		}
	}
	return 0, false
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
