package lhd

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
			glog.Infof("init lhd id:%s", t.DeskData.Game.Id)
			t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
			// glog.Infof("init lhd data:%#v", games)
			break
		}
	}

	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//下注时间
	t.BetTime = int(t.Game.LHD.LHBetTime)
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
	//龙虎斗的赔率
	t.DeskFree.LHDeskFree.OddsMap = make(map[uint32]int32)
	t.DeskFree.LHDeskFree.OddsMap[0] = 9
	t.DeskFree.LHDeskFree.OddsMap[1] = 2
	t.DeskFree.LHDeskFree.OddsMap[2] = 2
}

//.

// '进入房间响应消息
func (t *Desk) freeEnterMsg(userid string) *pb.LHFreeEnterRoomRsp {
	msg := new(pb.LHFreeEnterRoomRsp)
	//房间数据
	msg.Roominfo = handler.PackLHFreeRoom(t.DeskData)
	//坐下玩家信息
	msg.Userinfo = t.freeSeatBetsMsg()
	//排行榜前6玩家
	msg.Rank = t.freeRankMsg(20)
	t.freeRoomDataMsg(msg.Roominfo)
	return msg
}

func (t *Desk) freeRoomDataMsg(msg *pb.LHFreeRoom) {
	msg.State = t.state
	// msg.Dealer = t.DeskGame.DealerSeat
	// msg.Userid = t.DeskGame.Dealer
	// if t.DeskFree != nil && t.DeskFree.Carry > 0 {
	// 	msg.Carry = uint32(t.DeskFree.Carry)
	// }
	// msg.LeftDealerNum = t.leftDealerTimes()
	// msg.DealerNum = DealerTimes
	// p := t.getPlayer(t.DeskGame.Dealer)
	// if p != nil {
	// 	msg.Photo = p.GetPhoto()
	// }
	// 输赢记录20条
	msg.Winer = t.DeskFree.LHHistory
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

// 进入消息
func (t *Desk) freeCameinMsg(userid string) {
	msg := new(pb.LHFreeCameinNtf)
	msg.Userinfo = t.freeSeatRoleMsg(userid)
	t.broadcast4(msg)
}

// 位置上玩家数据
func (t *Desk) freeSeatRoleMsg(userid string) (msg *pb.LHFreeUser) {
	var user *data.User
	v := t.roles[userid]
	if v == nil {
		user = t.NewbiewRobot[userid]
	} else {
		user = v.User
	}
	msg = handler.PackLHFreeUser(user)
	if t.DeskFree == nil {
		return
	}
	// 胜场
	if w, ok := user.FreeWinMap[int32(pb.LHD)]; ok {
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
func (t *Desk) freeSeatBetsMsg() (msg []*pb.LHFreeUser) {
	for _, v := range t.roles {
		msg2 := t.freeSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	for k, u := range t.NewbiewRobot {
		usermsg := handler.PackLHFreeUser(u)
		if t.DeskFree == nil {
			continue
		}
		// 胜场
		if w, ok := u.FreeWinMap[int32(pb.LHD)]; ok {
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
func (t *Desk) freeRankMsg(num int32) (msg []*pb.LHFreeUser) {
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

// // 玩家下注数据
// func (t *Desk) freeBetsMsg(userid string) (msg []*pb.LHRoomBets) {
// 	for i := pb.DESK_SEAT2; i <= pb.DESK_SEAT4; i++ {
// 		seat := uint32(i)
// 		bets := t.getFreeSeatBet(userid, seat)
// 		msg2 := &pb.LHRoomBets{
// 			Seat: seat,
// 			Bets: bets,
// 		}
// 		msg = append(msg, msg2)
// 	}
// 	return
// }

// // 玩家位置下注数量
// func (t *Desk) getFreeSeatBet(userid string, seat uint32) int64 {
// 	if t.DeskFree == nil {
// 		return 0
// 	}
// 	if m, ok := t.SeatRoleBets[seat]; ok {
// 		return m[userid]
// 	}
// 	return 0
// }

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
// 				// msg := handler.LhdBeDealerMsg(0, int64(num), t.DeskFree.CarryInit, t.DeskGame.Dealer,
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
// func (t *Desk) dealerListMsg() (msg *pb.LHFreeDealerListRsp) {
// 	msg = new(pb.LHFreeDealerListRsp)
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
// 		msg2 := &pb.LHDealerList{
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
	if t.DeskFree.LHSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.LHSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		switch seatBet {
		case 0:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[2]) {
				return pb.BetTopLimit //tie下注限制
			}
		case 1:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[0]) {
				return pb.BetTopLimit //龙下注限制
			}
		case 2:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[1]) {
				return pb.BetTopLimit //虎下注限制
			}
		default:
			return pb.OperateError
		}
	}
	cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE75), fmt.Sprintf("lhd%s房间下注", t.DeskData.Rid))
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
	if t.DeskFree.LHSeatRoleBets[seatBet] != nil {
		c := t.DeskFree.LHSeatRoleBets[seatBet][userid]
		limit := c.GetSum() + num
		switch seatBet {
		case 0:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[2]) {
				return pb.BetTopLimit //tie下注限制
			}
		case 1:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[0]) {
				return pb.BetTopLimit //龙下注限制
			}
		case 2:
			if limit > int64(t.DeskData.Game.LHD.BetLimit[1]) {
				return pb.BetTopLimit //虎下注限制
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
	_, ntf := resFreeBet(seatBet, num,
		t.DeskFree.SeatBets[seatBet], betsNum, userid)
	t.broadcast(ntf)
	return pb.OK
}

// 下注消息
func resFreeBet(beseat uint32, val, coin,
	bets int64, userid string) (*pb.LHFreeBetRsp, *pb.LHFreeBetNtf) {
	rsp := &pb.LHFreeBetRsp{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	ntf := &pb.LHFreeBetNtf{
		Beseat: beseat,
		Value:  uint32(val),
		Coin:   coin,
		Bets:   bets,
		Userid: userid,
	}
	return rsp, ntf
}

//.

// ' 超时处理
func (t *Desk) freeTimeout() {
	switch t.state {
	case int32(pb.STATE_READY):
		t.freeStart()
	case int32(pb.STATE_BET):
		if t.timer == t.BetTime {
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
	var photo string
	var nickname string
	p := t.getPlayer(t.DeskGame.Dealer)
	if p != nil {
		photo = p.GetPhoto()
		nickname = p.GetNickname()
	}
	var left uint32 = t.leftDealerTimes()
	msg := resFreeStart(t.DeskGame.Dealer, photo, nickname,
		t.state, t.DeskFree.Carry, DealerTimes, left)
	t.broadcast(msg)
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
		// j := utils.RandIntN(i + 1)
		j := r.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
	t.DeskGame.Cards = d
}

//.

// '发牌
func (t *Desk) freeDeal(tie bool) {
	t.DeskFree.LHCards = make([]uint32, 0)
	for {
		cards := make([]uint32, 1)
		tmp := t.DeskGame.Cards[:1]
		copy(cards, tmp)
		if tie && len(t.DeskFree.LHCards) > 0 {
			// 要开和,根据第一张开的计算第二张开什么
			poker := t.DeskFree.LHCards[0]
			rank := algo.Rank(poker) // 点数
			suit := algo.Suit(poker) // 花色
			targetSuit := suit
			for i := 1; i <= int(utils.RandInt32N(3)+1); i++ {
				targetSuit -= 16
				if targetSuit == 0 {
					targetSuit = algo.Spade
				}
			}
			// 随一张其他花色的牌
			nextPoker := rank | targetSuit
			t.DeskFree.LHCards = append(t.DeskFree.LHCards, nextPoker)
			t.DeskGame.Cards = t.DeskGame.Cards[1:]
			return
		}
		if !tie && len(t.DeskFree.LHCards) > 0 {
			poker := t.DeskFree.LHCards[0]
			if algo.Rank(poker) == algo.Rank(cards[0]) {
				// 不开和的情况下开到和了重新开
				t.DeskGame.Cards = t.DeskGame.Cards[1:]
				continue
			}
		}
		t.DeskFree.LHCards = append(t.DeskFree.LHCards, cards[0])
		t.DeskGame.Cards = t.DeskGame.Cards[1:]
		if len(t.DeskFree.LHCards) >= 2 {
			break
		}
	}
}

//.

// '结束游戏
func (t *Desk) freeGameOver() {
	// t.shuffle()  //洗牌
	// t.freeDeal() //发牌
	// 结算
	// cs1 := t.getHandCards(uint32(pb.DESK_SEAT1)) //庄家牌
	// cs2 := t.getHandCards(uint32(pb.DESK_SEAT2)) //闲家牌
	//
	// var muliti2, muliti3, muliti4 int64 //庄家赢返回正数,输返回负数
	// var a1 uint32 = algo.Lhd(cs1, cs2)  //1位置庄家牌力,2位置闲家牌力
	// switch a1 {
	// case 0: //和
	// 	muliti2, muliti3, muliti4 = 1, 1, -8
	// case 1: //庄
	// 	muliti2, muliti3, muliti4 = -1, 1, 1
	// case 2: //闲
	// 	muliti2, muliti3, muliti4 = 1, -1, 1
	// }
	// t.DeskFree.Power[uint32(pb.DESK_SEAT1)] = algo.LhdRank(cs1)
	// t.DeskFree.Power[uint32(pb.DESK_SEAT2)] = algo.LhdRank(cs2)
	// //各位置和庄家对比的赔付倍数
	// t.DeskFree.Multiple[uint32(pb.DESK_SEAT2)] = muliti2
	// t.DeskFree.Multiple[uint32(pb.DESK_SEAT3)] = muliti3
	// t.DeskFree.Multiple[uint32(pb.DESK_SEAT4)] = muliti4
	//牌局数累加一次
	//t.DeskGame.Round++
	// t.xianjiaJiesuan() //结算,闲家赔付
	//庄家收钱,TODO 奖池抽成
	// t.dealerWin()
	//庄家赔付,闲家收钱,奖池抽成
	// t.dealerJiesuan()
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
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "30033")
	}
}

//.

// .结账
func (t *Desk) gameSettlement() []data.LHUserDetail {
	details := make([]data.LHUserDetail, 0)
	var cashStock, coinStock, mcashTaxStock, acashTaxStock, mcoinTaxStock, acoinTaxStock int64 = 0, 0, 0, 0, 0, 0 // 库存变化
	winner := t.getWinner()
	odds := t.DeskFree.LHDeskFree.OddsMap[winner] // 赔率
	roleBets := t.DeskFree.LHSeatRoleBets
	scoreMap := make(map[string]data.Currency)
	for k, roles := range roleBets {
		// 赢的玩家发奖
		for id, v := range roles {
			// var cash, coin, mcashTax, mcoinTax int64 = 0, 0, 0, 0
			role := t.roles[id]
			if role == nil {
				if !t.robotSettlement(id, k, winner, v) {
					// 新手状态人机结算
					glog.Errorf("lhd Settlement role is nil, id:%s", id)
				}
				continue
			}
			winBets := v.Scale(-1)
			if k == winner {
				w := v.Scale(float64(odds))
				winBets.Merge(w)
				// 返奖
				t.sendCurrency(id, w.Coin, w.Diamond, int32(pb.LOG_TYPE67), fmt.Sprintf("lhd房间%s结算", t.Game.Id))
			} else if winner == Tie {
				// 如果开了和，龙虎要退一半的钱
				bdiamond := int64(math.Floor(float64(v.Diamond) * 0.5))
				bcoin := int64(math.Ceil(float64(v.Coin) * 0.5))
				winBets.Merge(data.Currency{Diamond: bdiamond, Coin: bcoin})
				t.sendCurrency(id, bcoin, bdiamond, int32(pb.LOG_TYPE68), fmt.Sprintf("lhd房间%s开和退回", t.Game.Id))
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
			// t.sendCurrency(k, -mcoin, -mcash, int32(pb.LOG_TYPE69), fmt.Sprintf("lhd房间%s扣税", t.Game.Id))
		} else {
			// 扣暗税
			realCash = -int64(math.Max(float64(-realCash-role.GiveDiamond), 0))
			acash, acoin = t.getTax(realCash, c.Coin, false)
		}
		// acoin := t.getTax(c.Coin-mcoin, false)
		// 玩家输赢记录
		t.lhdRecord(k, c.GetSum())
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
				Gtype: uint32(pb.LHD),
				Win:   c.GetSum() > 0,
			}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType: int32(pb.LHD),
				Win:      c.GetSum() > 0,
			})

			// 对局详情
			details = append(details, t.createDetail(k, c, 0, 0, acash, acoin)) //, mcash, mcoin, acash, acoin))

			// 策略结算
			// t.strategySettlement(role.Userid)
		}
		// 记录输赢分
		// c.Merge(data.Currency{Diamond: -mcash, Coin: -mcoin})
		t.ScoreMap[k] = c

		if !role.GetRobot() {
			// 策略结算
			t.strategySettlement(role.Userid)
		}
	}
	for _, v := range t.LHObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			details = append(details, t.createDetail(v, data.Currency{}, 0, 0, 0, 0))
		}
	}

	// 游戏开奖记录
	t.addHistory(winner)
	// 库存变化
	t.changeStock(-cashStock, -coinStock, mcashTaxStock, mcoinTaxStock, acashTaxStock, acoinTaxStock, t.Game.Id)
	return details
}

// 输赢记录
func (t *Desk) lhdRecord(useid string, num int64) {
	bets := t.DeskFree.Bets[useid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.LHD), Score: num, Bet: bets}
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
	if w, ok := role.FreeWinMap[int32(pb.LHD)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.LHD)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	// 游戏局数
	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.LHD)]++

	// 下注
	role.LHDStrategy.RoundBet = append(role.LHDStrategy.RoundBet, bets)
	if len(role.LHDStrategy.RoundBet) > 50 {
		role.LHDStrategy.RoundBet = role.LHDStrategy.RoundBet[len(role.LHDStrategy.RoundBet)-50:]
	}

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    role.Pid,
			Userid:     role.Userid,
			Gtype:      int32(pb.LHD),
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
			glog.Error("publish user lhd bets error", err)
		}

		log := handler.GameFlowWaterLog(bets, int(pb.LHD), 0, len(role.RechargeTarge), useid)
		myactor.Logger().Tell(log)
	}
}

// 开奖记录
func (t *Desk) addHistory(win uint32) {
	t.DeskFree.LHHistory = append(t.DeskFree.LHHistory, int32(win))
	size := len(t.DeskFree.LHHistory)
	if size > 20 { // 只保留20条
		t.DeskFree.LHHistory = t.DeskFree.LHHistory[size-20:]
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
func (t *Desk) createDetail(userid string, score data.Currency, mcash, mcoin, acash, acoin int64) data.LHUserDetail {
	bet := t.Bets[userid]
	detail := data.LHUserDetail{
		Userid:  userid,
		Dragon:  t.getBetNum(userid, 1),
		Tiger:   t.getBetNum(userid, 2),
		Tie:     t.getBetNum(userid, 0),
		Win:     score.GetSum(),
		Result:  t.getWinReslut(score.GetSum()),
		Observe: bet <= 0,
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

// ' 发牌消息
func resDraw(seat uint32, state int32, cards []uint32) *pb.LHDrawNtf {
	return &pb.LHDrawNtf{
		Seat:  seat,
		State: state,
		Cards: cards,
	}
}

//.

// ' 游戏开始消息
func resFreeStart(dealer, photo, nickname string, state int32,
	carry int64, dealerNum, left uint32) *pb.LHFreeGamestartNtf {
	return &pb.LHFreeGamestartNtf{
		Dealer:        dealer,
		Photo:         photo,
		State:         state,
		Coin:          carry,
		DealerNum:     dealerNum,
		LeftDealerNum: left,
		Nickname:      nickname,
	}
}

//.

// ' 游戏结束消息
func (t *Desk) resOverFree() *pb.LHFreeGameoverNtf {
	msg := &pb.LHFreeGameoverNtf{
		State:    t.state,
		Winner:   int32(t.getWinner()),
		Userinfo: t.freeSeatBetsMsg(),
	}
	for k, v := range t.ScoreMap {
		bean := &pb.LHRoomScore{
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

	// var role *data.DeskRole
	// userid := t.GetOnlyOnePlayer()
	// if userid != "" {
	// 	role = t.roles[userid]
	// }
	// if t.Dtype == int32(pb.DESK_TYPE_NORMAL) || (role != nil && role.RegistArea == 0) {
	// 	// 自然概率
	// 	t.NatureWinner()
	// 	// 通知玩家
	// 	ntf := new(pb.LHFreeLeadNtf)
	// 	ntf.Data = t.LeadMsg()
	// 	t.broadcast(ntf)
	// 	return
	// }

	// 判断该哪边赢
	winner, err1 := t.calWinner(room)
	if err1 != nil {
		glog.Infof("%s", err1.Error())
		// 自然概率
		t.NatureWinner()
	} else {
		// 分牌
		t.dispensePoker(winner)
	}

	// 通知玩家
	ntf := new(pb.LHFreeLeadNtf)
	ntf.Data = t.LeadMsg()
	t.broadcast(ntf)
}

// 根据下注总额获取对应的库存
func (t *Desk) getRoom() (data.Game, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := config.GetGames()
	var roomId int32 = 0
	for _, g := range games {
		if g.Gtype != int32(pb.LHD) || g.Dtype != t.Dtype {
			continue
		}
		switch t.Dtype {
		case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
			roomId = g.LHD.ControlRoom[0]
		// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
		// 	roomId = g.LHD.NoviceRoom[0]
		case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
			int32(pb.DESK_TYPE_NEWBIEW): // 正常、新手
			intervals := g.LHD.BetInterval
			for i, v := range intervals {
				if pool >= int64(v[0]) && v[1] == -1 || (pool >= int64(v[0]) && pool < int64(v[1])) {
					roomId = g.LHD.BetRoom[i][0]
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
			room.LHD.NewbiewMode = room.LHD.ANewbiewMode
		}
	}
	return room, nil
}

// 计算哪边赢或者开和
func (t *Desk) calWinner(room data.Game) (uint32, error) {
	// lhtMap := make(map[uint32]int64)
	//防刷水检测
	if l, ok := t.preventionCheat(room); ok {
		return l, nil
	}
	// 配置结果
	// if env == "dev" {
	// 	w := t.cfgWinner()
	// 	if w >= 0 {
	// 		// if w == 0 {
	// 		// 	t.freeDeal(true)
	// 		// }
	// 		return uint32(w), nil
	// 	}
	// }
	// B类策略
	// reslut, trigger := t.WinStrategy_B()
	// if trigger {
	// 	if reslut == Tie {
	// 		t.freeDeal(true)
	// 	}
	// 	return reslut, nil
	// }

	// 选择策略
	t.selectStrategy()
	for _, s := range t.LHDeskFree.LHDStrategys {
		t.LHDeskFree.TriggerStrategy = s.Id
		if len(s.Seat) > 0 {
			seat, _ := utils.ChoiceInt(s.Seat)
			glog.Infof("lhd winner seat %d, strategy:%d, gameid:%s", seat, s.Id, t.GameId)
			return uint32(seat), nil
		} else {
			// 自然概率
			glog.Infof("lhd no seat, strategy:%d, gameid:%s", s.Id, t.GameId)
			return 0, fmt.Errorf("use nature")
		}
	}

	// 自然概率
	return 0, fmt.Errorf("use nature")

	// // 当前总下注
	// betPool := t.DeskGame.CashBets
	// // 最终系数
	// factor := t.getFinalFactor()
	// if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
	// 	// 点控
	// 	factor = t.getControlFactor()
	// }
	// t.DeskGame.FinaFactor = factor
	// glog.Infof("lhd FinalFactor %f, stockid:%s, rid:%s, gameid:%s", factor, t.Game.Id, t.Rid, t.GameId)
	// // 输分上限
	// maxLose := t.getMaxLose()
	// // maxLose := room.LHD.LoseLimited
	// // 平台胜率(万分比)
	// winPro := room.LHD.Winning[len(room.LHD.FinalCoefficient)]
	// for i, v := range room.LHD.FinalCoefficient {
	// 	if factor <= float64(v)/100 {
	// 		winPro = room.LHD.Winning[i]
	// 		break
	// 	}
	// }
	// // 新手房间走独立的胜率
	// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 	winPro = 10000 - t.getNewBiewWinning()
	// }
	// var lorh uint32 = uint32(utils.RandInt32N(2) + 1)

	// // 哪方赢
	// playerwin := false
	// // 开和的输赢
	// tieWin := t.getTieWin()
	// tempMap := make(map[uint32]int64)
	// if utils.RandInt32N(10000)+1 <= winPro {
	// 	glog.Infof("lhd player lose, winPro %d stockid:%s, rid:%s, gameid:%s", winPro, t.Game.Id, t.Rid, t.GameId)
	// 	// 平台赢
	// 	// 计算开什么赢得更多
	// 	// var winBet int64 = 0
	// 	for i := 1; i < 3; i++ {
	// 		betCash := t.DeskFree.SeatCashBets[uint32(i)]
	// 		odds := t.DeskFree.LHDeskFree.OddsMap[uint32(i)] // 赔率
	// 		win := betPool - int64(odds)*betCash
	// 		if win >= 0 {
	// 			lhtMap[uint32(i)] = win
	// 		}
	// 	}
	// 	if tieWin > 0 {
	// 		lhtMap[0] = tieWin
	// 	}
	// } else {
	// 	glog.Infof("lhd player win, winPro %d stockid:%s, rid:%s, gameid:%s", (10000 - winPro), t.Game.Id, t.Rid, t.GameId)
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
	// 				lhtMap[uint32(i)] = win
	// 			}
	// 			if win >= 0 {
	// 				tempMap[uint32(i)] = win
	// 			}
	// 		} else {
	// 			//
	// 			if win < 0 && (math.Abs(float64(win)) <= float64(maxLose) || maxLose == 0) {
	// 				lhtMap[uint32(i)] = win
	// 			}
	// 			if win >= 0 {
	// 				tempMap[uint32(i)] = win
	// 			}
	// 		}
	// 	}
	// 	if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) {
	// 		if tieWin < 0 && t.maxOutCash(-tieWin) {
	// 			lhtMap[Tie] = tieWin
	// 		}
	// 	} else {
	// 		if tieWin < 0 && (math.Abs(float64(tieWin)) <= float64(maxLose) || maxLose == 0) {
	// 			lhtMap[Tie] = tieWin
	// 		}
	// 	}

	// 	// 如果三个都是平台赢
	// 	if len(lhtMap) <= 0 {
	// 		lhtMap = tempMap
	// 	}
	// 	// if _, ok := lhtMap[Tie]; ok && len(lhtMap) == 1 {
	// 	// 	// 当只能开和的时候
	// 	// }
	// }
	// // 计算开哪个位置
	// lorh = t.choiceWinner(lhtMap, tempMap, playerwin)
	// if lorh == 0 {
	// 	t.freeDeal(true)
	// }
	// return lorh, nil
}

// 计算哪边赢或者开和
// func (t *Desk) calWinnerNewBiew(room data.Game) (uint32, error) {
// 	beginner := table.GetTables().NewbieTable.Get()
// 	if beginner.Id == 0 {
// 		return 0, errors.New("[lhd] no find beginner config")
// 	}
// 	// 哪方赢
// 	playerwin := false
// 	tempMap := make(map[uint32]int64)
// 	interval := beginner.OutCashInterval
// 	winning := interval[len(interval)-1]
// 	for userid, r := range t.roles {
// 		for i := 1; i < len(interval); i++ {
// 			if r.OutDiamond < int64(interval[i]) {
// 				winning = beginner.Winning[i-1]
// 				break
// 			}
// 		}
// 		if utils.RandWan(winning) {
// 			// 玩家赢
// 			playerwin = true

// 		}
// 	}

// 	return 1, nil
// }

// 分配牌
func (t *Desk) dispensePoker(lorh uint32) {
	// 发牌
	t.freeDeal(lorh == Tie)
	if lorh == 0 {
		t.Power[1] = t.DeskFree.LHCards[0]
		t.Power[2] = t.DeskFree.LHCards[1]
		return
	}
	// 把大牌给赢的一方
	if algo.Rank(t.DeskFree.LHCards[0]) > algo.Rank(t.DeskFree.LHCards[1]) {
		t.Power[lorh] = t.DeskFree.LHCards[0]
		if lorh == 1 {
			t.Power[2] = t.DeskFree.LHCards[1]
		} else {
			t.Power[1] = t.DeskFree.LHCards[1]
		}
	} else {
		t.Power[lorh] = t.DeskFree.LHCards[1]
		if lorh == 1 {
			t.Power[2] = t.DeskFree.LHCards[0]
		} else {
			t.Power[1] = t.DeskFree.LHCards[0]
		}
	}
}

// 开牌结果msg
func (t *Desk) LeadMsg() []*pb.LHFreeRoomOver {
	lead := make([]*pb.LHFreeRoomOver, 0)
	for k, v := range t.Power {
		msg := new(pb.LHFreeRoomOver)
		msg.Ptype = int32(k)
		msg.Value = v
		lead = append(lead, msg)
	}
	return lead
}

// 玩家下注总额
func (t *Desk) userSeatBetMsg(userid string) []*pb.LHFreeUserBet {
	bets := make([]*pb.LHFreeUserBet, 0)
	for k, v := range t.DeskFree.LHSeatRoleBets {
		value := v[userid]
		msg := new(pb.LHFreeUserBet)
		msg.Ptype = int32(k)
		msg.Bets = uint32(value.GetSum())
		bets = append(bets, msg)
	}
	return bets
}

// .赢家
func (t *Desk) getWinner() uint32 {
	if algo.Rank(t.Power[1]) > algo.Rank(t.Power[2]) {
		return 1
	} else if algo.Rank(t.Power[1]) < algo.Rank(t.Power[2]) {
		return 2
	}
	return 0
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
			msg := new(pb.LHFreeRankNtf)
			msg.Rank = rank
			t.broadcast4(msg)
		}
		return
	}
	msg := new(pb.LHFreeRankNtf)
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
		glog.Errorf("lhd event Marshal fail")
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
		// 新手房没下注的不记
		return
	}
	// de := data.LHDetail{
	// 	Winner:      t.getWinner(),
	// 	DragonValue: algo.GetSuitString(t.Power[1]) + algo.GetRankString(t.Power[1]),
	// 	TigerValue:  algo.GetSuitString(t.Power[2]) + algo.GetRankString(t.Power[2]),
	// 	Bets:        t.CashBets + t.CoinBets,
	// 	UserDetail:  details,
	// 	StrategyId:  t.LHDeskFree.TriggerStrategy,
	// }
	t.LHDetail.Winner = t.getWinner()
	t.LHDetail.DragonValue = algo.GetSuitString(t.Power[1]) + algo.GetRankString(t.Power[1])
	t.LHDetail.TigerValue = algo.GetSuitString(t.Power[2]) + algo.GetRankString(t.Power[2])
	t.LHDetail.Bets = t.CashBets + t.CoinBets
	t.LHDetail.StrategyId = t.LHDeskFree.TriggerStrategy
	t.LHDetail.UserDetail = details

	var playerWin int64 = 0
	for _, ld := range details {
		playerWin += ld.Win
	}
	t.LHDetail.PlayerWin = playerWin
	// 保存详情
	t.saveDetail()
}

// 保存详情
func (t *Desk) saveDetail() {
	detail := data.Detail{
		WaterId:      t.GameId,
		BeginTime:    t.BeginTime,
		EndTime:      utils.BsonNow().Unix(),
		Gtype:        int32(pb.LHD),
		RoomId:       t.Rid,
		DeskId:       t.Game.Id,
		PlayerFactor: int(t.DeskGame.FinaFactor * 100),
		LHDetail:     &t.LHDetail,
	}
	players := make([]string, 0)
	for _, ld := range t.LHDetail.UserDetail {
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
		if r.OutDiamond+out <= int64(t.Game.LHD.NewbiewMode.OutCashLimited) {
			return true
		}
	}
	return false
}

// 新手人机结算
func (t *Desk) robotSettlement(id string, seat, winner uint32, c data.Currency) bool {
	if user, ok := t.NewbiewRobot[id]; ok {
		if seat == winner {
			odds := t.DeskFree.LHDeskFree.OddsMap[winner] // 赔率
			user.AddCoin(c.GetSum() * int64(odds))
			t.ScoreMap[id] = c.Scale(float64(odds))
		} else {
			// 记录输赢分
			if winner == Tie {
				c = c.Scale(0.5)
				user.AddCoin(c.GetSum())
				t.ScoreMap[id] = c.Scale(-0.5)
			} else {
				t.ScoreMap[id] = c.Scale(-1)
			}
		}
		return true
	}
	return false
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
