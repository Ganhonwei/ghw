package up7

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
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
	// 		glog.Infof("init UP id:%s", t.DeskData.Game.Id)
	// 		t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
	// 		// glog.Infof("init UP data:%#v", games)
	// 		break
	// 	}
	// }

	up7Room := table.GetTables().Up7RoomTable.Get()

	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//下注时间
	t.BetTime = int(up7Room.BetTime) //int(t.Game.UP.LHBetTime)
	//赠送金库存
	t.DeskFree.GiveStock = 0
	//userid:num, 玩家下注金额
	t.DeskFree.Bets = make(map[string]int64)
	//seat:num, 位置下注金额
	t.DeskFree.SeatBets = make(map[uint32]int64)
	//seat:num, 位置彩金下注金额
	t.DeskFree.SeatCashBets = make(map[uint32]int64)
	// 玩家下注记录
	t.DeskFree.SeatBetLogs = make(map[string][]data.Currency)
	t.BackOutDiamond = make(map[string]int64)
	t.DeskFree.LHDeskFree = &data.LHDeskFree{
		//位置下注详细
		LHSeatRoleBets: make(map[uint32]map[string]data.Currency),
		//人机下注位置
		LHRobotSeat: make(map[string]uint32),
		LHRobotBets: make(map[string]int64),
		// 防刷水局数
		LHCheat:  make(map[string]int32),
		ScoreMap: make(map[string]data.Currency),
	}
	//玩家个人系数
	t.DeskFree.UserFactorMap = make(map[string]float64)

	t.state = int32(pb.STATE_READY)
	//7up的赔率
	t.DeskFree.LHDeskFree.OddsMap = make(map[uint32]int32)
	// for _, m := range table.GetTables().Up7MultipleTable.GetDataList() {
	// 	if m.IsDefault == 1 {
	// 		t.DeskFree.LHDeskFree.OddsMap[uint32(m.Point)] = m.Multiple
	// 	}
	// }

	for _, m := range table.GetTables().Up7MultipleTableBing.GetDataList() {
		t.DeskFree.LHDeskFree.OddsMap[uint32(m.BetArea)] = m.ProbItem[0].Multiple + 1
	}
	// 暴击赔率
	t.DeskFree.LHDeskFree.CritOddsMap = make(map[uint32]int32)

	// 检查对局人机
	t.checkNewbiewRobot()
}

// 检查增加或减少假人
func (t *Desk) checkNewbiewRobot() {
	robotNum := int(table.GetTables().Up7RobotTable.Get().Num)
	initScore := table.GetTables().Up7RobotTable.Get().InitScore

	more := robotNum - len(t.NewbiewRobot)
	if more == 0 {
		return
	} else if more > 0 {
		// 生成假人
		for i := range more {
			user := &data.User{
				Nickname:   login.RandName(),
				Userid:     fmt.Sprintf("%d", handler.GenerateOrderId(uint32(i))),
				Photo:      strconv.Itoa(utils.RandIntN(30) + 1),
				Coin:       int64(utils.RandMN(int(initScore[0]), int(initScore[1]))),
				FreeWinMap: make(map[int32][]data.FreeWin),
				Robot:      true,
			}
			// 人机vip头像
			handler.RobotVip(user)
			// 自定义头像
			t.getHead(user)
			user.FreeWinMap[int32(pb.SEVEN)] = handler.GenNewbiewFreeWin(int32(pb.SEVEN))
			t.NewbiewRobot[user.Userid] = user
		}
	} else if more < 0 {
		// 减少家人
		less := -more

		var kickRobots []string
		for k := range t.NewbiewRobot {
			kickRobots = append(kickRobots, k)
			if len(kickRobots) >= less {
				break
			}
		}
		for _, k := range kickRobots {
			if u, ok := t.NewbiewRobot[k]; ok {
				// 还头像
				t.returnHead(u)
				delete(t.NewbiewRobot, k)

				//离开消息
				t.broadcast(&pb.CRASHLeaveNtf{Userid: k})
			}
		}
	}
}

// '进入房间响应消息
func (t *Desk) freeEnterMsg(userid string) *pb.UPFreeEnterRoomRsp {
	msg := new(pb.UPFreeEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackUPFreeRoom(t.DeskData)
	t.freeRoomDataMsg(userid, msg.Roominfo)
	//坐下玩家信息
	msg.Userinfo = t.freeSeatBetsMsg()
	//排行榜前20玩家
	msg.Rank = t.freeRankMsg(20)
	return msg
}

func (t *Desk) freeRoomDataMsg(userid string, msg *pb.UPFreeRoom) {
	var registArea int
	if r, ok := t.roles[userid]; ok {
		registArea = r.RegistArea
	}

	msg.State = t.state

	// 筹码
	up7 := table.GetTables().Up7Table.Get(1)
	msg.ChipSeat = up7.ChipDefault[registArea]
	for _, chip := range up7.Chips[registArea].Value {
		msg.Chip = append(msg.Chip, uint32(chip))
	}
	for _, chip := range up7.ChipShowDefault[registArea].Value {
		msg.ChipShowDefault = append(msg.ChipShowDefault, uint32(chip))
	}
	if bets, ok := t.DeskFree.Bets[userid]; ok {
		msg.PlayerBets = bets
	}
	msg.BetLimit = up7.BetLimit[registArea].Value
	msg.Odds = t.LHDeskFree.OddsMap
	msg.CritOdds = t.LHDeskFree.CritOddsMap

	// 输赢记录20条
	msg.Point = t.DeskFree.LHHistory
	// 输赢记录100条
	msg.Histories = t.DeskFree.UPHistory
	msg.Dice = t.LHCards
	msg.Players = int32(len(t.DeskFree.Bets))

	tt := 0
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		switch t.state {
		case int32(pb.STATE_READY):
			tt = FreeReadyTime - t.timer
		case int32(pb.STATE_DEALING):
			tt = FreeDealTime - t.timer
		case int32(pb.STATE_BET):
			tt = t.BetTime - t.timer
		case int32(pb.STATE_LEAD):
			tt = FreeLeadTime - t.timer
			// 开牌阶段要把开的牌发下去
			// for _, v := range t.LHCards {
			// 	msg.Dice = append(msg.Dice, v)
			// }
		case int32(pb.STATE_OVER):
			tt = FreeSettlementTime - t.timer
			// 结算阶段要把开的牌发下去
			// for _, v := range t.LHCards {
			// 	msg.Dice = append(msg.Dice, v)
			// }
		default:
			tt = t.BetTime - t.timer
		}
	}
	if tt < 0 {
		tt = 0
	}
	msg.Timer = uint32(tt)
}

// 进入消息
func (t *Desk) freeCameinMsg(userid string) {
	msg := new(pb.UPFreeCameinNtf)
	msg.Userinfo = t.freeSeatRoleMsg(userid)
	t.broadcast4(msg)
}

// 位置上玩家数据
func (t *Desk) freeSeatRoleMsg(userid string) (msg *pb.UPFreeUser) {
	var user *data.User
	v := t.roles[userid]
	if v == nil {
		user = t.NewbiewRobot[userid]
	} else {
		user = v.User
	}
	msg = handler.PackUPFreeUser(user)
	if t.DeskFree == nil {
		return
	}
	// 胜场
	if w, ok := user.FreeWinMap[int32(pb.SEVEN)]; ok {
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
func (t *Desk) freeSeatBetsMsg() (msg []*pb.UPFreeUser) {
	for _, v := range t.roles {
		msg2 := t.freeSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	for k, u := range t.NewbiewRobot {
		usermsg := handler.PackUPFreeUser(u)
		if t.DeskFree == nil {
			continue
		}
		// 胜场
		if w, ok := u.FreeWinMap[int32(pb.SEVEN)]; ok {
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
func (t *Desk) freeRankMsg(num int32) (msg []*pb.UPFreeUser) {
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

// 玩家下注数据
func (t *Desk) freeBetsMsg(userid string) (msg []*pb.UPRoomBets) {
	for i := pb.DESK_SEAT2; i <= pb.DESK_SEAT4; i++ {
		seat := uint32(i)
		bets := t.getFreeSeatBet(userid, seat)
		msg2 := &pb.UPRoomBets{
			Seat: seat,
			Bets: bets,
		}
		msg = append(msg, msg2)
	}
	return
}

// 玩家位置下注数量
func (t *Desk) getFreeSeatBet(userid string, seat uint32) int64 {
	if t.DeskFree == nil {
		return 0
	}
	if m, ok := t.SeatRoleBets[seat]; ok {
		return m[userid]
	}
	return 0
}

//.

// 'beDealer 没人上庄时都可以选择上庄,已经上庄的人可以补庄
//
//st:0下庄 1上庄 2补庄
// func (t *Desk) beDealer(userid string, st int32, num uint32) pb.ErrCode {
// 	if !t.isFree() {
// 		return pb.NotDealerRoom
// 	}
// 	user := t.getPlayer(userid)
// 	if user == nil {
// 		glog.Errorf("userid %s not exist", userid)
// 		return pb.NotInRoom
// 	}
// 	//TODO 全部带上庄
// 	//num = uint32(user.GetCoin())
// 	switch st {
// 	case int32(pb.DEALER_DOWN):
// 		switch t.state {
// 		case int32(pb.STATE_BET): //下注中
// 			if userid == t.DeskGame.Dealer {
// 				t.DeskFree.DealerDown = true
// 				// msg := handler.UPBeDealerMsg(0, int64(num), t.DeskFree.CarryInit, t.DeskGame.Dealer,
// 				// 	userid, user.GetNickname(), user.GetPhoto())
// 				// msg.Down = true
// 				// t.broadcast(msg)
// 				return pb.OK
// 			}
// 			return pb.DealerDownFail
// 		default:
// 			t.delBeDealer(userid, user)
// 			return pb.OK
// 		}
// 	case int32(pb.DEALER_UP):
// 		//已经上庄,暂时不能重复上
// 		if t.alreadyBeDealer(userid) {
// 			return pb.BeDealerAlready
// 		}
// 		//上庄限制
// 		if num < t.DeskData.Carry {
// 			return pb.BeDealerNotEnough
// 		}
// 		t.addBeDealer(userid, st, int64(num), user)
// 		return pb.OK
// 	case int32(pb.DEALER_BU):
// 		t.addBeDealer(userid, st, int64(num), user)
// 		return pb.OK
// 	}
// 	return pb.OperateError
// }

//.

// '获取上庄列表消息
// func (t *Desk) dealerListMsg() (msg *pb.UPFreeDealerListRsp) {
// 	msg = new(pb.UPFreeDealerListRsp)
// 	if !t.isFree() {
// 		msg.Error = pb.NotDealerRoom
// 		return
// 	}
// 	for k, v := range t.DeskFree.Dealers {
// 		user := t.getPlayer(k)
// 		if user == nil {
// 			glog.Errorf("userid %s not exist", k)
// 			continue
// 		}
// 		msg2 := &pb.UPDealerList{
// 			Userid:   k,
// 			Nickname: user.GetNickname(),
// 			Photo:    user.GetPhoto(),
// 			//Coin:     user.GetCoin(),
// 			Coin: v,
// 		}
// 		msg.List = append(msg.List, msg2)
// 	}
// 	return
// }

//.

// 玩家倍投下注(当前下注每一门注码都加倍)
func (t *Desk) freeDoubleBet(userid string) (rsp *pb.UPDoubleBetRsp) {
	rsp = new(pb.UPDoubleBetRsp)
	rsp.Error = pb.OK

	role, ok := t.roles[userid]
	if !ok {
		glog.Errorf("userid %s not exist", userid)
		rsp.Error = pb.NotInRoom
		return
	}
	if t.state != int32(pb.STATE_BET) {
		rsp.Error = pb.BetOver
		return
	}

	status := table.GetTables().Up7RoomTable.Get().Status
	if status == 0 {
		rsp.Error = pb.RoomMaintenance
		return
	}

	num := t.DeskFree.Bets[userid]
	if num <= 0 {
		rsp.Error = pb.NoBet
		return
	}
	if !t.canBet(role, num) {
		rsp.Error = pb.NotEnoughCoin
		return
	}

	// 下注限制
	betLimit := table.GetTables().Up7Table.Get(1).BetLimit[role.RegistArea]
	if num < betLimit.Value[0] || num > betLimit.Value[1] {
		rsp.Error = pb.BetTopLimit //下注限制
		return
	}
	if bets, ok := t.DeskFree.Bets[userid]; ok {
		if bets+num > betLimit.Value[1] {
			rsp.Error = pb.BetTopLimit //下注限制
			return
		}
	}

	// cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	// coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	// t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE76), fmt.Sprintf("UP%s房间下注", t.DeskData.Rid))
	// 投注加倍
	var newBets, newCashBet, newCoinBet int64
	var seatValue, seatBets, seatPools = make(map[uint32]int64), make(map[uint32]int64), make(map[uint32]int64)
	for seatBet, m := range t.LHSeatRoleBets {
		if c, ok := m[userid]; ok {
			bet := c.GetSum()
			cashBet := t.getCashOrCoinBet(role, bet, true) // 彩金下注
			coinBet := bet - cashBet                       // 奖励金下注

			c.Diamond += cashBet
			c.Coin += coinBet
			m[userid] = c
			t.LHSeatRoleBets[seatBet] = m

			newCashBet += cashBet
			newCoinBet += coinBet

			newBets += bet
			seatValue[seatBet] = bet
			seatBets[seatBet] = c.GetSum()
			seatPools[seatBet] = t.DeskFree.SeatBets[seatBet] + bet

			//下注记录
			t.DeskFree.Bets[userid] += bet              //个人总下注额
			t.DeskFree.SeatBets[seatBet] += bet         //当前位置总下注额(没扣税的)
			t.DeskFree.SeatCashBets[seatBet] += cashBet //当前位置彩金总下注额
			t.DeskFree.SeatBetLogs[userid] = append(t.DeskFree.SeatBetLogs[userid], data.Currency{
				Diamond: cashBet,
				Coin:    coinBet,
				Tag:     seatBet,
			})
		}
	}
	t.sendCurrency(userid, (-1 * newCoinBet), (-1 * newCashBet), int32(pb.LOG_TYPE76), fmt.Sprintf("UP%s房间下注", t.DeskData.Rid))

	user := role.User
	if !user.GetRobot() {
		t.DeskGame.CashBets += newCashBet //当局彩金总下注额
		t.DeskGame.CoinBets += newCoinBet //当局奖励金总下注额
		if user.OutDiamond-user.Diamond > 0 {
			// 下注时使用了可提现彩金需要记录
			t.BackOutDiamond[userid] = user.OutDiamond - user.Diamond
		}
	}
	role.NoBetTimes = 0 // 不下注次数清零

	rsp.Userid = user.Userid
	rsp.Value = newBets
	rsp.Bets = t.DeskFree.Bets[userid]
	rsp.SeatValue = seatValue
	rsp.SeatBets = seatBets
	rsp.SeatPools = seatPools
	return
}

// 撤销上一步下注
func (t *Desk) freeUndoBet(userid string) (rsp *pb.UPUndoBetRsp) {
	rsp = new(pb.UPUndoBetRsp)
	rsp.Error = pb.OK

	_, ok := t.roles[userid]
	if !ok {
		glog.Errorf("userid %s not exist", userid)
		rsp.Error = pb.NotInRoom
		return
	}

	if t.state != int32(pb.STATE_BET) {
		rsp.Error = pb.BetOver
		return
	}

	betLogs := t.DeskFree.SeatBetLogs[userid]
	if len(betLogs) == 0 {
		rsp.Error = pb.NoBet
		return
	}

	bet := betLogs[len(betLogs)-1]
	t.sendCurrency(userid, bet.Coin, bet.Diamond, int32(pb.LOG_TYPE151), fmt.Sprintf("UP房间%s撤销下注", t.DeskData.Rid))
	t.DeskFree.SeatBetLogs[userid] = betLogs[:len(betLogs)-1]

	seatBet := bet.Tag
	num := bet.GetSum()

	//下注记录
	t.DeskFree.Bets[userid] -= num                  //个人总下注额
	t.DeskFree.SeatBets[seatBet] -= num             //当前位置总下注额(没扣税的)
	t.DeskFree.SeatCashBets[seatBet] -= bet.Diamond //当前位置彩金总下注额

	if bets := t.DeskFree.Bets[userid]; bets <= 0 {
		delete(t.DeskFree.Bets, userid)
	}

	var playerSeatBets int64 // 该玩家撤销位置的当前下注额
	if m, ok := t.LHSeatRoleBets[seatBet]; ok {
		if c, ok := m[userid]; ok {
			c.Coin -= bet.Coin
			c.Diamond -= bet.Diamond
			if c.Coin < 0 {
				c.Coin = 0
				glog.Errorf("up undo bet coin error: undo=%#v, currency=%#v", bet, c)
			}
			if c.Diamond < 0 {
				c.Diamond = 0
				glog.Errorf("up undo bet diamond error: undo=%#v, currency=%#v", bet, c)
			}
			if c.GetSum() <= 0 {
				delete(m, userid)
			} else {
				m[userid] = c
				playerSeatBets = c.GetSum()
			}
			t.LHSeatRoleBets[seatBet] = m
		}
	}
	rsp.UndoBets = num
	rsp.BetSeat = seatBet
	rsp.SeatBets = playerSeatBets
	rsp.SeatPool = t.DeskFree.SeatBets[seatBet]
	rsp.Bets = t.DeskFree.Bets[userid]
	rsp.Players = int32(len(t.DeskFree.Bets))
	return
}

// '百人下注
func (t *Desk) freeBet(userid string, seatBet uint32, num int64) pb.ErrCode {
	if t.state != int32(pb.STATE_BET) {
		return pb.BetOver
	}
	if num <= 0 {
		return pb.OperateError
	}
	status := table.GetTables().Up7RoomTable.Get().Status
	if status == 0 {
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
	if !t.canBet(role, num) {
		return pb.NotEnoughCoin
	}

	// 下注限制
	betLimit := table.GetTables().Up7Table.Get(1).BetLimit[role.RegistArea]
	if num < betLimit.Value[0] || num > betLimit.Value[1] {
		return pb.BetTopLimit //下注限制
	}
	if bets, ok := t.DeskFree.Bets[userid]; ok {
		if bets+num > betLimit.Value[1] {
			return pb.BetTopLimit //下注限制
		}
	}

	// 下注位置有误
	if odd, ok := t.LHDeskFree.OddsMap[seatBet]; !ok || odd <= 0 {
		return pb.OperateError
	}

	otherSeatBet := seatBet
	if seatBet == 0 {
		otherSeatBet = 1
	} else if seatBet == 1 {
		otherSeatBet = 0
	}
	if t.DeskFree.SeatCashBets[otherSeatBet] > 0 {
		return pb.OperateError
	}

	cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE76), fmt.Sprintf("UP%s房间下注", t.DeskData.Rid))
	//下注记录
	t.DeskFree.Bets[userid] += num              //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num         //当前位置总下注额(没扣税的)
	t.DeskFree.SeatCashBets[seatBet] += cashBet //当前位置彩金总下注额
	t.DeskFree.SeatBetLogs[userid] = append(t.DeskFree.SeatBetLogs[userid], data.Currency{
		Diamond: cashBet,
		Coin:    coinBet,
		Tag:     seatBet,
	})

	if !user.GetRobot() {
		t.DeskGame.CashBets += cashBet //当局彩金总下注额
		t.DeskGame.CoinBets += coinBet //当局奖励金总下注额
		if user.OutDiamond-user.Diamond > 0 {
			// 下注时使用了可提现彩金需要记录
			t.BackOutDiamond[userid] = user.OutDiamond - user.Diamond
		}
	}
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.LHSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Diamond += cashBet
		c.Coin += coinBet
		m[userid] = c
		t.LHSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Diamond: cashBet,
			Coin:    coinBet,
		}
		t.LHSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	role.NoBetTimes = 0 // 不下注次数清零
	msg, ntf := t.resFreeBet(seatBet, cashBet+coinBet,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.send2userid(userid, msg)
	t.broadcast2(userid, ntf)
	return pb.OK
}

func (t *Desk) newBiewFreeBet(userid string, seatBet uint32, num int64) pb.ErrCode {
	if t.state != int32(pb.STATE_BET) {
		return pb.BetOver
	}
	status := table.GetTables().Up7RoomTable.Get().Status
	if status == 0 {
		return pb.RoomMaintenance
	}
	user := t.NewbiewRobot[userid]
	if user == nil {
		glog.Errorf("newbiew robot userid %s not exist", userid)
		return pb.NotInRoom
	}
	// 下注限制
	betLimit := table.GetTables().Up7Table.Get(1).BetLimit[0]
	if bets, ok := t.DeskFree.Bets[userid]; ok {
		if bets+num > betLimit.Value[1] {
			return pb.BetTopLimit //下注限制
		}
	}

	// if t.DeskFree.LHSeatRoleBets[seatBet] != nil {
	// 	c := t.DeskFree.LHSeatRoleBets[seatBet][userid]
	// 	limit := c.GetSum() + num
	// 	switch seatBet {
	// 	case 0:
	// 		if limit > int64(t.DeskData.Game.UP.BetLimit[2]) {
	// 			return pb.BetTopLimit //tie下注限制
	// 		}
	// 	case 1:
	// 		if limit > int64(t.DeskData.Game.UP.BetLimit[0]) {
	// 			return pb.BetTopLimit //龙下注限制
	// 		}
	// 	case 2:
	// 		if limit > int64(t.DeskData.Game.UP.BetLimit[1]) {
	// 			return pb.BetTopLimit //虎下注限制
	// 		}
	// 	default:
	// 		return pb.OperateError
	// 	}
	// }
	user.AddCoin(-num)
	t.DeskFree.Bets[userid] += num      //个人总下注额
	t.DeskFree.SeatBets[seatBet] += num //当前位置总下注额(没扣税的)
	//位置详细记录
	var betsNum int64 //玩家当前位置下注总数
	if m, ok := t.LHSeatRoleBets[seatBet]; ok {
		c := m[userid]
		c.Coin += num
		m[userid] = c
		t.LHSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	} else {
		m := make(map[string]data.Currency)
		m[userid] = data.Currency{
			Coin: num,
		}
		t.LHSeatRoleBets[seatBet] = m
		betsNum = m[userid].GetSum()
	}
	_, ntf := t.resFreeBet(seatBet, num,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.broadcast(ntf)
	return pb.OK
}

// 下注消息
func (t *Desk) resFreeBet(beseat uint32, val, coin,
	bets int64, userid string) (*pb.UPFreeBetRsp, *pb.UPFreeBetNtf) {
	players := int32(len(t.DeskFree.Bets))
	rsp := &pb.UPFreeBetRsp{
		Beseat:  beseat,
		Value:   uint32(val),
		Coin:    coin,
		Bets:    bets,
		Userid:  userid,
		Players: players,
	}
	ntf := &pb.UPFreeBetNtf{
		Beseat:  beseat,
		Value:   uint32(val),
		Coin:    coin,
		Bets:    bets,
		Userid:  userid,
		Players: players,
	}
	return rsp, ntf
}

//.

//'游戏状态

//.

//.

// '发牌
// func (t *Desk) freeDeal(tie bool) {
// 	t.DeskFree.LHCards = make([]uint32, 0)
// 	for {
// 		point := utils.RandInt32N(6) + 1
// 		if tie && len(t.DeskFree.LHCards) > 0 {
// 			// 总点数为7
// 			firstPoint := t.DeskFree.LHCards[0] // 第一个骰子点数
// 			t.DeskFree.LHCards = append(t.DeskFree.LHCards, 7-firstPoint)
// 			return
// 		}
// 		if !tie && len(t.DeskFree.LHCards) > 0 {
// 			firstPoint := t.DeskFree.LHCards[0]
// 			if firstPoint+uint32(point) == 7 {
// 				// 不开7的情况下开到7了重新开
// 				continue
// 			}
// 		}
// 		t.DeskFree.LHCards = append(t.DeskFree.LHCards, uint32(point))
// 		if len(t.DeskFree.LHCards) >= 2 {
// 			break
// 		}
// 	}
// }

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
	//记录房间上局赢家
	// t.saveWiners()
	//个人记录
	// t.setFreeRecord()
	//机器人操作
	t.kickRobot()
	//踢除离线玩家
	t.kickOffline()
	//剔除5局不下注的玩家
	// t.kickNoBet()
	//剔除点控房间多余玩家
	// t.kickPCMoreUser()
	//踢出超时或余额不足的人
	t.limitOver()
	//消息广播
	// t.freeStart()
	//奖池清零
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//赠送金库存变化
	if t.GiveStock > 0 {
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "40033")
	}
}

// .结账
func (t *Desk) gameSettlement() []data.LHUserDetail {
	details := make([]data.LHUserDetail, 0)
	// var cashStock, coinStock, mcashTaxStock, acashTaxStock, mcoinTaxStock, acoinTaxStock int64 = 0, 0, 0, 0, 0, 0 // 库存变化
	winner, point := t.getWinner()
	// 赔率
	var winnerOdds, pointOdds int32
	if winner != 7 {
		if critOdds, ok := t.LHDeskFree.CritOddsMap[winner]; ok {
			winnerOdds = critOdds // 暴击
		} else {
			winnerOdds = t.LHDeskFree.OddsMap[winner]
		}
	}
	if critOdds, ok := t.LHDeskFree.CritOddsMap[point]; ok {
		pointOdds = critOdds // 暴击
	} else {
		pointOdds = t.LHDeskFree.OddsMap[point]
	}

	scoreMap := make(map[string]data.Currency)
	roleBets := t.DeskFree.LHSeatRoleBets
	for k, roles := range roleBets {
		// 赢的玩家发奖
		for id, v := range roles {
			role := t.roles[id]
			if role == nil {
				if !t.robotSettlement(id, k, winner, point, v, winnerOdds, pointOdds) {
					// 新手状态人机结算
					glog.Errorf("7up Settlement role is nil, id:%s", id)
				}
				continue
			}
			winBets := v.Scale(-1)
			// 压大小
			if k == winner && winner != 7 {
				w := v.Scale(float64(winnerOdds))
				winBets.Merge(w)
				// 得到的彩金和奖励金
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE70), fmt.Sprintf("7up房间%s结算", t.Rid))
			}
			// 压点数
			if k == point {
				w := v.Scale(float64(pointOdds))
				winBets.Merge(w)
				// 得到的彩金和奖励金
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE70), fmt.Sprintf("7up房间%s结算", t.Rid))
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
		role := t.roles[k]
		// 玩家输赢记录
		t.upRecord(k, c.GetSum())
		// 非机器人记录
		if !role.GetRobot() {
			// 事件(非机器人)
			bean := &event.GameRecordEvent{
				Gtype: uint32(pb.SEVEN),
				Win:   c.GetSum() > 0,
			}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType: int32(pb.SEVEN),
				Win:      c.GetSum() > 0,
			})
			// 对局详情
			details = append(details, t.createDetail(k, c, 0, 0, 0, 0))
		}
		// 记录输赢分
		t.ScoreMap[k] = c

		// if !role.GetRobot() {
		// 	// 策略结算
		// 	t.strategySettlement(role.Userid)
		// }
	}
	for _, v := range t.LHObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			details = append(details, t.createDetail(v, data.Currency{}, 0, 0, 0, 0))
		}
	}

	// 游戏开奖记录
	t.addHistory()
	// 库存变化(算上明税暗税)
	// changeCash := t.DeskGame.CashBets - (cashStock + mcashTaxStock + acashTaxStock)
	// changeCoin := t.DeskGame.CoinBets - (coinStock + mcoinTaxStock + acoinTaxStock)
	// t.changeStock(-cashStock, -coinStock, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock, t.Game.Id)
	return details
}

// 输赢记录
func (t *Desk) upRecord(useid string, num int64) {
	bets := t.DeskFree.Bets[useid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.SEVEN), Score: num, Bet: bets}
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

	// 修改玩家数据
	wins := make([]data.FreeWin, 0)
	if role.FreeWinMap == nil {
		role.FreeWinMap = make(map[int32][]data.FreeWin)
	}
	if w, ok := role.FreeWinMap[int32(pb.SEVEN)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.SEVEN)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.SEVEN)]++

	// 下注
	role.SevenStrategy.RoundBet = append(role.SevenStrategy.RoundBet, bets)
	if len(role.SevenStrategy.RoundBet) > 50 {
		role.SevenStrategy.RoundBet = role.SevenStrategy.RoundBet[len(role.SevenStrategy.RoundBet)-50:]
	}

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    role.Pid,
			Userid:     role.Userid,
			Gtype:      int32(pb.SEVEN),
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
			glog.Error("publish user seven bets error", err)
		}
		log := handler.GameFlowWaterLog(bets, int(pb.SEVEN), 0, len(role.RechargeTarge), useid)
		myactor.Logger().Tell(log)
	}
}

// 开奖记录
func (t *Desk) addHistory() {
	t.DeskFree.LHHistory = append(t.DeskFree.LHHistory, int32(t.LHCards[0]+t.LHCards[1]))
	size := len(t.DeskFree.LHHistory)
	if size > 20 { // 只保留20条
		t.DeskFree.LHHistory = t.DeskFree.LHHistory[size-20:]
	}

	winner, point := t.getWinner()
	t.DeskFree.UPHistory = append(t.DeskFree.UPHistory, &pb.UPHistory{
		Winner: winner,
		Point:  point,
		Crit:   len(t.LHDeskFree.CritOddsMap) > 0,
		Lead:   []uint32{t.LHCards[0], t.LHCards[1]},
	})
	size = len(t.DeskFree.UPHistory)
	if size > 100 {
		t.DeskFree.UPHistory = t.DeskFree.UPHistory[size-100:]
	}
}

// . 对局详情
func (t *Desk) createDetail(userid string, score data.Currency, mcash, mcoin, acash, acoin int64) data.LHUserDetail {
	bet := t.Bets[userid]
	detail := data.LHUserDetail{
		Userid:   userid,
		Dragon:   t.getBetNum(userid, 0),
		Tiger:    t.getBetNum(userid, 1),
		Tie:      t.getBetNum(userid, 7),
		SeatBets: make(map[uint32]int64),
		Win:      score.GetSum(),
		Result:   t.getWinReslut(score.GetSum()),
		Observe:  bet <= 0,
	}
	for seat, m := range t.DeskFree.LHSeatRoleBets {
		if bets, ok := m[userid]; ok {
			detail.SeatBets[seat] = bets.GetSum()
			if _, ok := t.DeskFree.CritOddsMap[seat]; ok {
				detail.Crit++
			}
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
	if role.Money > 0 {
		detail.BeforeBackRate = int((int64(role.CashOut) + detail.BeforeCash) * 10000 / int64(role.Money))
		detail.AfterBackRate = int((int64(role.CashOut) + role.Diamond) * 10000 / int64(role.Money))
	}
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

//.

// ' 游戏结束消息
func (t *Desk) resOverFree() *pb.UPFreeGameoverNtf {
	winner, point := t.getWinner()
	msg := &pb.UPFreeGameoverNtf{
		State:       t.state,
		Winner:      int32(winner),
		WinnerPoint: int32(point),
		Userinfo:    t.freeSeatBetsMsg(),
		Crit:        len(t.DeskFree.CritOddsMap) > 0,
		Lead:        t.DeskFree.LHCards,
	}
	for k, v := range t.ScoreMap {
		bean := &pb.UPRoomScore{
			Userid: k,
		}
		bean.Score = v.GetSum()
		msg.Scores = append(msg.Scores, bean)
		// 跑马灯检测
		t.checkMarquee(k, v)
		// 赠送金变化
		t.checkGiveDiamond(k, v)
		// 打码量
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
	// if t.tickStop {
	// 	return
	// }
	t.timer++
	//看看下注时间到没有
	switch t.state {
	case int32(pb.STATE_READY):
		// 空闲状态
		if t.timer >= FreeReadyTime {
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
			// 下注结束
			t.timer = 0
			nextState := int32(pb.STATE_LEAD) // 开牌状态
			during := int32(1)
			// 判断暴击
			if crit := t.critMultiples(); crit {
				nextState = int32(pb.STATE_DECLARE)
				during = 2
			}

			t.state = nextState
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = nextState
			msg.During = during
			t.selfPid.Tell(msg)
			// t.tickStop = true
			return
		}
	case int32(pb.STATE_DECLARE):
		// 暴击展示状态
		if t.timer >= FreeCritTime {
			critMsg := &pb.UPCritMultipleNtf{
				Odds:     t.LHDeskFree.OddsMap,
				CritOdds: t.LHDeskFree.CritOddsMap,
			}
			t.broadcast(critMsg)

			t.timer = 0
			t.state = int32(pb.STATE_LEAD) // 开牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_LEAD)
			msg.During = 1
			t.selfPid.Tell(msg)
		}
		// t.tickStop = true

	case int32(pb.STATE_LEAD):
		// 开牌状态
		if t.timer >= FreeLeadTime {
			t.timer = 0
			// t.state = int32(pb.STATE_OVER) // 结算状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_OVER)
			msg.During = FreeSettlementTime
			t.selfPid.Tell(msg)
			// t.tickStop = true
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

// 暴击开奖倍数
// func (t *Desk) critMultiples() (crit bool) {
// 	multiples := table.GetTables().Up7MultipleTable.GetDataList()
// 	pointMultiples := make(map[int32][]utils.ChoiceI64[int32])
// 	for _, m := range multiples {
// 		c := utils.ChoiceI64[int32]{Item: m.Id}
// 		c.Weight = parsePctMultiple(m.Rate, 100_000_000_000)
// 		pointMultiples[m.Point] = append(pointMultiples[m.Point], c)
// 	}

// 	for seat, choices := range pointMultiples {
// 		c, err := utils.WeightedChoiceI64(choices)
// 		if err != nil {
// 			glog.Errorf("choice crit multiple error:", err)
// 			continue
// 		}
// 		multiple := table.GetTables().Up7MultipleTable.Get(c.Item)
// 		if multiple == nil {
// 			glog.Errorf("get crit multiple error:", c.Item)
// 			continue
// 		}
// 		// 随到默认倍数跳过
// 		if multiple.IsDefault == 1 {
// 			continue
// 		}
// 		crit = true
// 		// 暴击倍数
// 		t.LHDeskFree.CritOddsMap[uint32(seat)] = multiple.Multiple
// 	}
// 	return
// }

// 暴击开奖倍数(饼)
func (t *Desk) critMultiples() (crit bool) {
	multiples := table.GetTables().Up7MultipleTableBing.GetDataList()
	pointMultiples := make(map[int32][]utils.ChoiceI64[int32])
	for _, m := range multiples {
		for _, p := range m.ProbItem {
			c := utils.ChoiceI64[int32]{Item: p.Multiple, Weight: int64(math.Floor(p.Prob * 100000))}
			pointMultiples[m.BetArea] = append(pointMultiples[m.BetArea], c)
		}
	}

	for betArea, choices := range pointMultiples {
		base := choices[0].Item + 1
		c, err := utils.WeightedChoiceI64(choices)
		if err != nil {
			glog.Errorf("choice crit multiple error:", err)
			continue
		}
		multiple := c.Item + 1
		// 随到默认倍数跳过
		if multiple == base {
			continue
		}
		crit = true
		// 暴击倍数
		t.LHDeskFree.CritOddsMap[uint32(betArea)] = multiple
	}
	return
}

// parsePctMultiple 解析百分比概率
func parsePctMultiple(pct string, base float64) int64 {
	if pct == "" {
		return 0
	}

	if strings.HasSuffix(pct, "%") {
		r1 := strings.ReplaceAll(pct, "%", "")
		r2, _ := strconv.ParseFloat(r1, 64)
		return int64(r2 * float64(base) / 100)
	}
	r2, _ := strconv.ParseFloat(pct, 64)
	return int64(r2 * float64(base))
}

// 开牌
func (t *Desk) lead() {
	// room, err := t.getRoom()
	// if err != nil {
	// 	// 异常牌局
	// 	glog.Errorf("err room")
	// 	return
	// }
	// t.Game = room

	// var role *data.DeskRole
	// userid := t.GetOnlyOnePlayer()
	// if userid != "" {
	// 	role = t.roles[userid]
	// }
	// if t.Dtype == int32(pb.DESK_TYPE_NORMAL) || (role != nil && role.RegistArea == 0) {
	// 	// 自然概率
	// 	t.NatureWinner()
	// 	// 通知玩家
	// 	ntf := new(pb.UPFreeLeadNtf)
	// 	ntf.Data = t.LHCards
	// 	t.broadcast(ntf)
	// 	return
	// }
	// 摇骰子
	// t.freeDeal(false)
	// 判断该哪边赢
	// err1 := t.calWinner(room)
	// 自然概率
	t.NatureWinner()

	// 通知玩家
	ntf := new(pb.UPFreeLeadNtf)
	ntf.Data = t.LHCards
	t.broadcast(ntf)
}

// 根据下注总额获取对应的库存
func (t *Desk) _getRoom() (data.Game, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := config.GetGames()
	var roomId int32 = 0
	for _, g := range games {
		if g.Gtype != int32(pb.SEVEN) || g.Dtype != t.Dtype {
			continue
		}
		switch t.Dtype {
		case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
			roomId = g.UP.ControlRoom[0]
		// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
		// 	roomId = g.UP.NoviceRoom[0]
		case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
			int32(pb.DESK_TYPE_NEWBIEW): // 正常
			intervals := g.UP.BetInterval
			for i, v := range intervals {
				if pool >= int64(v[0]) && v[1] == -1 || (pool >= int64(v[0]) && pool < int64(v[1])) {
					roomId = g.UP.BetRoom[i][0]
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
			room.UP.NewbiewMode = room.UP.ANewbiewMode
		}
	}
	return room, nil
}

// 计算哪边赢或者开和
func (t *Desk) _calWinner(room data.Game) error {
	//防刷水检测
	if _, ok := t.preventionCheat(room); ok {
		return nil
	}

	// 选择策略
	t._selectStrategy()
	for _, s := range t.LHDeskFree.LHDStrategys {
		t.LHDeskFree.TriggerStrategy = s.Id
		if len(s.Seat) > 0 {
			seat, _ := utils.ChoiceInt(s.Seat)
			glog.Infof("7up winner seat %d, strategy:%d, gameid:%s", seat, s.Id, t.GameId)
			t.dice(uint32(seat)) //摇骰子
			return nil
		} else {
			// 自然概率
			glog.Infof("7up no seat, strategy:%d, gameid:%s", s.Id, t.GameId)
			return fmt.Errorf("use nature")
		}
	}

	// 自然概率
	return fmt.Errorf("use nature")
	// 配置结果
	// w := t.cfgWinner()
	// if w >= 0 {
	// 	t.dice(uint32(w)) //摇骰子
	// 	return nil
	// }
	// B类策略
	// reslut, trigger := t.WinStrategy_B()
	// if trigger {
	// 	t.dice(reslut) //摇骰子
	// 	glog.Infof("UP player win trigger strategy, stockid:%s, rid:%s, gameid:%s", t.Game.Id, t.Rid, t.GameId)
	// 	return nil
	// }

	// // 当前总下注
	// betPool := t.DeskGame.CashBets
	// // 最终系数
	// factor := t.getFinalFactor()
	// if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
	// 	// 点控
	// 	factor = t.getControlFactor()
	// }
	// t.DeskGame.FinaFactor = factor

	// glog.Infof("UP FinalFactor %f, stockid:%s, rid:%s, gameid:%s", factor, t.Game.Id, t.Rid, t.GameId)
	// // 输分上限
	// maxLose := t.getMaxLose()
	// // maxLose := room.UP.LoseLimited
	// // 平台胜率(万分比)
	// winPro := room.UP.Winning[len(room.UP.FinalCoefficient)]
	// for i, v := range room.UP.FinalCoefficient {
	// 	if factor <= float64(v)/100 {
	// 		winPro = room.UP.Winning[i]
	// 		break
	// 	}
	// }
	// // 新手房间走独立的胜率
	// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 	winPro = 10000 - t.getNewBiewWinning()
	// }
	// var lorh uint32 = uint32(utils.RandInt32N(2) + 1)
	// // tie := t.DeskFree.SeatCashBets[0]
	// // odds := oddsMap[0]

	// // 哪方赢
	// playerwin := false
	// // 开7的输赢
	// tieWin := t.getTieWin()
	// tempMap := make(map[uint32]int64)
	// if utils.RandInt32N(10000)+1 <= winPro {
	// 	glog.Infof("UP player lose, winPro %d  stockid:%s, rid:%s, gameid:%s", winPro, t.Game.Id, t.Rid, t.GameId)
	// 	// 平台赢
	// 	// 计算开什么赢得更多
	// 	for i := 1; i < 3; i++ {
	// 		betCash := t.DeskFree.SeatCashBets[uint32(i)]
	// 		odds := t.LHDeskFree.OddsMap[uint32(i)] // 赔率
	// 		win := betPool - int64(odds)*betCash
	// 		if win >= 0 {
	// 			resultMap[uint32(i)] = win
	// 		}
	// 		// if win > winBet {
	// 		// 	winBet = win
	// 		// 	lorh = uint32(i)
	// 		// }
	// 	}
	// 	if tieWin > 0 {
	// 		resultMap[0] = tieWin
	// 	}
	// 	// 如果开7大或小赢的一样，有3/52概率开到7
	// 	// if tieWin > 0 && rand.Int31n(52)+1 <= 3 {
	// 	// 	// 有3/52概率开出和
	// 	// 	lorh = 0
	// 	// 	// t.freeDeal(true)
	// 	// }
	// } else {
	// 	glog.Infof("UP player win, winPro %d stockid:%s, rid:%s, gameid:%s", (10000 - winPro), t.Game.Id, t.Rid, t.GameId)
	// 	// 玩家赢
	// 	playerwin = true
	// 	if tieWin >= 0 {
	// 		tempMap[Tie] = tieWin
	// 	}
	// 	for i := 1; i < 3; i++ {
	// 		v := t.DeskFree.SeatCashBets[uint32(i)]
	// 		odds := t.DeskFree.LHDeskFree.OddsMap[uint32(i)] // 赔率
	// 		win := betPool - int64(odds)*v
	// 		if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 			if win < 0 && t.maxOutCash(-win) {
	// 				resultMap[uint32(i)] = win
	// 			}
	// 			if win >= 0 {
	// 				tempMap[uint32(i)] = win
	// 			}
	// 		} else {
	// 			if win < 0 && (math.Abs(float64(win)) <= float64(maxLose) || maxLose == 0) {
	// 				resultMap[uint32(i)] = win
	// 			}
	// 			if win >= 0 {
	// 				tempMap[uint32(i)] = win
	// 			}
	// 		}
	// 	}

	// 	if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 		if tieWin < 0 && t.maxOutCash(-tieWin) {
	// 			resultMap[Tie] = tieWin
	// 		}
	// 	} else {
	// 		if tieWin < 0 && (math.Abs(float64(tieWin)) <= float64(maxLose) || maxLose == 0) {
	// 			resultMap[Tie] = tieWin
	// 		}
	// 	}
	// 	// 如果三个都是平台赢
	// 	if len(resultMap) <= 0 {
	// 		resultMap = tempMap
	// 	}
	// }
	// lorh = t.choiceWinner(resultMap, tempMap, playerwin)
	// t.dice(lorh) //摇骰子
	// glog.Infof("UP winner seat %d, gameid:%s, rid:%s, tieWin:%d", lorh, t.GameId, t.Rid, tieWin)
	// return nil
}

// 分配牌
// func (t *Desk) dispensePoker(lorh uint32) {
// 	if lorh == 0 {
// 		t.Power[1] = t.DeskFree.LHCards[0]
// 		t.Power[2] = t.DeskFree.LHCards[1]
// 		return
// 	}
// 	// 把大牌给赢的一方
// 	if algo.Rank(t.DeskFree.LHCards[0]) > algo.Rank(t.DeskFree.LHCards[1]) {
// 		t.Power[lorh] = t.DeskFree.LHCards[0]
// 		if lorh == 1 {
// 			t.Power[2] = t.DeskFree.LHCards[1]
// 		} else {
// 			t.Power[1] = t.DeskFree.LHCards[1]
// 		}
// 	} else {
// 		t.Power[lorh] = t.DeskFree.LHCards[1]
// 		if lorh == 1 {
// 			t.Power[2] = t.DeskFree.LHCards[0]
// 		} else {
// 			t.Power[1] = t.DeskFree.LHCards[0]
// 		}
// 	}
// }

// 开牌结果msg
func (t *Desk) LeadMsg() *pb.UPFreeRoomBets {
	msg := new(pb.UPFreeRoomBets)
	for k, v := range t.SeatBets {
		msg := new(pb.UPFreeRoomBets)
		msg.Ptype = int32(k)
		msg.Bets = v
	}
	return msg
}

// 玩家下注总额
func (t *Desk) userSeatBetMsg(userid string) []*pb.UPFreeUserBet {
	bets := make([]*pb.UPFreeUserBet, 0)
	for k, v := range t.DeskFree.LHSeatRoleBets {
		value := v[userid]
		msg := new(pb.UPFreeUserBet)
		msg.Ptype = int32(k)
		msg.Bets = uint32(value.GetSum())
		bets = append(bets, msg)
	}
	return bets
}

// .赢家
// winner: 0小,1大,7和
// point: 赢点数
func (t *Desk) getWinner() (winner uint32, point uint32) {
	point = t.DeskFree.LHCards[0] + t.DeskFree.LHCards[1]
	if point < 7 {
		winner = 0
	} else if point > 7 {
		winner = 1
	} else if point == 7 {
		winner = 7
	}
	return
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
			msg := new(pb.UPFreeRankNtf)
			msg.Rank = rank
			t.broadcast4(msg)
		}
		return
	}
	msg := new(pb.UPFreeRankNtf)
	msg.Rank = rank
	t.broadcast4(msg)
}

// 防刷水
func (t *Desk) preventionCheat(room data.Game) (uint32, bool) {

	return 0, false
}

// 事件
func (t *Desk) eventPost(userid string, eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("UP event Marshal fail")
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
func (t *Desk) gameDetail(details []data.LHUserDetail) {
	if len(details) <= 0 {
		// 房间里没人的不记
		return
	}

	winner, point := t.getWinner()
	t.UPDetail.Winner = winner
	t.UPDetail.PointValue = int32(point)
	t.UPDetail.Bets = t.CashBets + t.CoinBets
	t.UPDetail.StrategyId = t.LHDeskFree.TriggerStrategy
	t.UPDetail.Crit7up = true
	t.UPDetail.Crited = int16(len(t.DeskFree.CritOddsMap))
	t.UPDetail.CritOdds = t.DeskFree.CritOddsMap
	t.UPDetail.UserDetail = details
	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	t.UPDetail.PlayerWin = playerWin
	// 保存详情
	t.saveDetail(t.UPDetail)
	glog.Debugf("player win score %d, gameId:%s, seat:%d, point=%d", playerWin, t.GameId, winner, point)
}

// 保存详情
func (t *Desk) saveDetail(de data.UPDetail) {
	detail := data.Detail{
		WaterId:      t.GameId,
		BeginTime:    t.BeginTime,
		EndTime:      utils.BsonNow().Unix(),
		Gtype:        int32(pb.SEVEN),
		RoomId:       t.Rid,
		DeskId:       t.Game.Id,
		PlayerFactor: int(t.DeskGame.FinaFactor * 100),
		UPDetail:     &de,
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

// 剔除点控房间多余玩家
func (t *Desk) kickPCMoreUser() {
	if t.Dtype != int32(pb.DESK_TYPE_POINTCONTROL) {
		return
	}
	r, _ := t.roleCountNum()
	if r <= 1 {
		return
	}
	for k, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		errcode := t.leave(k, 0)
		if errcode != pb.OK {
			continue
		}
		glog.Infof("pc control kick user:%s, rid:%s", k, t.Rid)
		// 通知自己回到大厅
		ntf := &pb.UPLeaveNtf{
			Userid: k,
		}
		t.send2userid(k, ntf)
		//离开状态消息
		t.userLeaveDesk(k)
		t.offlineMsg(k)
		r--
		if r <= 1 {
			break
		}
	}
}

func (t *Desk) maxOutCash(out int64) bool {
	if t.Dtype != int32(pb.DESK_TYPE_NEWBIEW) {
		return false
	}
	// bean := table.GetTables().NewbieTable.Get()
	for _, r := range t.roles {
		if r.OutDiamond+out <= int64(t.Game.UP.NewbiewMode.OutCashLimited) {
			return true
		}
	}
	return false
}

// 新手人机结算
func (t *Desk) robotSettlement(id string, seat, winner, point uint32, c data.Currency, winnerOdds, pointOdds int32) bool {
	if user, ok := t.NewbiewRobot[id]; ok {
		winBets := c.Scale(-1)

		if seat == winner && winner != 7 {
			w := c.Scale(float64(winnerOdds))
			user.AddCoin(w.GetSum())
		}
		if seat == point {
			w := c.Scale(float64(pointOdds))
			user.AddCoin(w.GetSum())
		}

		// 记录输赢分
		if c, ok := t.ScoreMap[id]; ok {
			c.Merge(winBets)
			t.ScoreMap[id] = c
		} else {
			t.ScoreMap[id] = winBets
		}
		return true
	}
	return false
}

func (t *Desk) GameTime(userid string) {
	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		msg := &pb.LogGameTime{
			Userid: userid,
			Gtype:  int32(pb.SEVEN),
			Time:   time.Now().Unix() - t.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2userid(userid, msg)
	}
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
			// 玩家只要和下注超过20就开和
			if sbet, ok := t.LHSeatRoleBets[Tie]; ok {
				if sbet[role.Userid].GetSum() == bet {
					t.send2userid(role.Userid, &pb.TriggerFreeWelfare{Userid: role.Userid})
					return Tie, true
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
