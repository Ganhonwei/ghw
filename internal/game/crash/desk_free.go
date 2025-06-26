package crash

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
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

var r *rand.Rand

// '初始化
func (t *Desk) freeInit() {
	// 初始化配置的时候可能还没同步过来, 确保有数据
	// for {
	// 	games := config.GetGames()
	// 	if len(games) > 0 {
	// 		glog.Infof("init crash id:%s", t.DeskData.Game.Id)
	// 		t.DeskData.Game = config.GetGame(t.DeskData.Game.Id)
	// 		break
	// 	}
	// }
	if t.DeskFree.CRASHDeskFree == nil {
		t.DeskFree.CRASHDeskFree = new(data.CRASHDeskFree)
		t.CRASHDeskFree.CrashAutoLeave = make(map[string]int32)
		t.CRASHDeskFree.CrashAutoLeave1 = make(map[string]int32)
		t.CRASHDeskFree.CrashNextBets = make(map[string]data.Currency)
		t.CRASHDeskFree.CrashNextBets1 = make(map[string]data.Currency)
		t.CRASHDeskFree.LastRecord = new(pb.CrashLastRoundRecord)
	}
	t.DeskGame.GameId = handler.GenGameId(t.Gtype) // 生成对局号
	t.DeskGame.BeginTime = utils.BsonNow().Unix()  // 开始时间
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//下注时间
	room := table.GetTables().CrashRoomTable.Get()
	t.BetTime = int(room.BetTime)
	//userid:num, 玩家下注金额
	t.DeskFree.Bets = make(map[string]int64)
	//seat:num, 位置下注金额
	t.DeskFree.SeatBets = make(map[uint32]int64)
	//seat:num, 位置彩金下注金额
	t.DeskFree.SeatCashBets = make(map[uint32]int64)
	//结算需要退的可提现彩金
	t.BackOutDiamond = make(map[string]int64)
	//彩金变化库存
	t.DeskFree.CRASHCashStock = 0
	t.DeskFree.GiveStock = 0
	//奖励金变化库存
	t.DeskFree.CRASHCoinStock = 0
	t.CRASHMCashTax = 0
	t.CRASHMCoinTax = 0
	t.CRASHACashTax = 0
	t.CRASHACoinTax = 0
	t.GiveStock = 0
	//玩家个人系数
	t.DeskFree.UserFactorMap = make(map[string]float64)
	//下注信息
	t.DeskFree.CrashBets = make(map[string]data.Currency)
	t.DeskFree.CrashBets1 = make(map[string]data.Currency)
	t.DeskFree.CRASHScoreMap = make(map[string]data.Currency)
	t.DeskFree.CRASHScoreMap1 = make(map[string]data.Currency)
	t.DeskFree.CRASHTaxMap = make(map[string]int64)
	// 撤离信息
	t.DeskFree.CRASHBack = make(map[string]int32)
	t.DeskFree.CRASHBack1 = make(map[string]int32)
	t.CRASHPlayerJackpot = make(map[string]int64)
	// 打码排行榜
	t.CRASHDeskFree.BetRank = t.CRASHDeskFree.BetRank[:0]
	//下注倍数
	t.CRASHMulpitle = 100
	// t.CRASHControlMulpitle = 0
	//已赔付
	t.DeskFree.CRASHLose = 0
	//清空观察者
	t.CRASHObserver = make([]string, 0)
	//清空策略
	t.CRASHStrategys = make([]data.StrategyInfo, 0)
	t.CRASHTriggerStrategys = make(map[int]data.StrategyInfo)
	t.CRASHFirstPartIn = false
	//切换为准备状态
	t.state = int32(pb.STATE_READY)
	//对局详情
	t.CRASHDetail = data.Detail{
		WaterId:     t.DeskGame.GameId,
		BeginTime:   utils.BsonNow().Unix(),
		Gtype:       int32(pb.CRASH),
		RoomId:      t.Rid,
		CRASHDetail: &data.CRASHDetail{},
	}

	// 检查对局人机
	t.checkNewbiewRobot()
}

// 检查增加或减少假人
func (t *Desk) checkNewbiewRobot() {
	robotNum := int(table.GetTables().CrashRobotTable.Get().Num)
	initScore := table.GetTables().CrashRobotTable.Get().InitScore

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
			user.FreeWinMap[int32(pb.CRASH)] = handler.GenNewbiewFreeWin(int32(pb.CRASH))
			t.NewbiewRobot[user.Userid] = user
		}
	} else if more < 0 {
		// 减少假人
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
func (t *Desk) freeEnterMsg(userid string) *pb.CRASHEnterRoomRsp {
	msg := new(pb.CRASHEnterRoomRsp)
	//房间数据
	// msg.Roominfo = handler._PackCrashFreeRoom(t.DeskData)
	msg.Roominfo = t.packCrashFreeRoom(userid)
	// 下注排行榜
	msg.BetRank = t.packFreeBetRank()
	t.freeRoomDataMsg(msg.Roominfo, userid)
	//坐下玩家信息
	msg.Userinfo = t.freeSeatBetsMsg()
	//排行榜前6玩家
	// msg.Rank = t.freeRankMsg(20)
	// 自动逃离
	if v, ok := t.CRASHDeskFree.CrashAutoLeave[userid]; ok {
		msg.AutoCrash = true
		msg.CrashMultiple = v
	}
	if v, ok := t.CRASHDeskFree.CrashAutoLeave1[userid]; ok {
		msg.AutoCrash1 = true
		msg.CrashMultiple1 = v
	}
	return msg
}

func (t *Desk) packCrashFreeRoom(userid string) (msg *pb.CRASHFreeRoom) {
	var registArea int
	if role := t.roles[userid]; role != nil {
		registArea = role.RegistArea
	}

	room := table.GetTables().CrashRoomTable.Get()
	crash0 := table.GetTables().CrashTable.Get(1)
	crash1 := table.GetTables().CrashTable.Get(2)

	msg = &pb.CRASHFreeRoom{
		Roomid: t.DeskData.Rid,     //牌局id
		Gtype:  t.DeskData.Gtype,   //game type
		Rtype:  t.DeskData.Rtype,   //room type
		Dtype:  t.DeskData.Dtype,   //desk type
		Rname:  t.DeskData.Rname,   //room name
		Count:  uint32(room.Count), //当前房间限制玩家数量
		// Chip:     t.DeskData.Game.CRASH.ChipLimit, //房间下注筹码
		// ChipSeat: int32(d.Game.CRASH.ChipSeat),    //筹码位置
	}
	for i, chip := range crash0.Chips[registArea].Value {
		msg.Chip = append(msg.Chip, uint32(chip))
		if chip == crash0.ChipDefault[registArea] {
			msg.ChipSeat = int32(i)
		}
	}
	for i, chip := range crash1.Chips[registArea].Value {
		msg.Chip1 = append(msg.Chip1, uint32(chip))
		if chip == crash1.ChipDefault[registArea] {
			msg.ChipSeat1 = int32(i)
		}
	}
	msg.AutoCrashDefaultMulpitle0 = int32(crash0.AutoBackDefaultMulpitle[registArea] * 100)
	msg.AutoCrashDefaultMulpitle1 = int32(crash1.AutoBackDefaultMulpitle[registArea] * 100)

	autoLimit0 := crash0.AutoBackMulpitleLimit[registArea]
	msg.AutoCrashMultipleLimit0 = []int32{
		utils.CaseElse(autoLimit0.Value[0] < 0, -1, int32(autoLimit0.Value[0]*100)),
		utils.CaseElse(autoLimit0.Value[1] < 0, -1, int32(autoLimit0.Value[1]*100)),
	}
	autoLimit1 := crash1.AutoBackMulpitleLimit[registArea]
	msg.AutoCrashMultipleLimit1 = []int32{
		utils.CaseElse(autoLimit1.Value[0] < 0, -1, int32(autoLimit1.Value[0]*100)),
		utils.CaseElse(autoLimit1.Value[1] < 0, -1, int32(autoLimit1.Value[1]*100)),
	}

	msg.BetLimit0 = crash0.BetLimit[registArea].Value
	msg.BetLimit1 = crash1.BetLimit[registArea].Value
	msg.BetDefault0 = int64(crash0.BetDefault[registArea])
	msg.BetDefault1 = int64(crash1.BetDefault[registArea])

	msg.WinLimit0 = crash0.WinLimit[registArea]
	msg.WinLimit1 = crash1.WinLimit[registArea]
	return
}

func (t *Desk) freeRoomDataMsg(msg *pb.CRASHFreeRoom, userid string) {
	msg.State = t.state
	// 输赢记录50条
	msg.History = t.DeskFree.CRASHHistory
	msg.Jackpot = t.CrashJackpot
	for _, v := range t.Bets {
		msg.Bets += v
	}
	if v, ok := t.CrashBets[userid]; ok {
		msg.Bets0 = v.GetSum()
	}
	if v, ok := t.CrashBets1[userid]; ok {
		msg.Bets1 = v.GetSum()
	}
	if v, ok := t.CrashNextBets[userid]; ok {
		msg.NextBets0 = v.GetSum()
	}
	if v, ok := t.CrashNextBets1[userid]; ok {
		msg.NextBets1 = v.GetSum()
	}
	msg.Players = int32(len(t.DeskFree.Bets)) // int32(len(t.roles) + len(t.NewbiewRobot))

	tt := 0
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		switch t.state {
		case int32(pb.STATE_READY):
			tt = 1 - t.timer/20
		case int32(pb.STATE_BET):
			tt = t.BetTime - t.timer/20
		case int32(pb.STATE_LEAD):
			msg.Multiple = t.CRASHMulpitle
			if v, ok := t.CRASHBack[userid]; ok {
				msg.Back = true
				msg.Backwin = t.CRASHScoreMap[userid].GetSum()
				msg.BackMultiple = v
			}
			if v, ok := t.CRASHBack1[userid]; ok {
				msg.Back1 = true
				msg.Backwin1 = t.CRASHScoreMap1[userid].GetSum()
				msg.BackMultiple1 = v
			}
			msg.TakeoffMs = t.CRASHTakeoffTime.UnixMilli()
			msg.CurMs = time.Now().UnixMilli()
		case int32(pb.STATE_OVER):
			tt = FreeSettlementTime - t.timer/20
			if v, ok := t.CRASHBack[userid]; ok {
				msg.Back = true
				msg.Backwin = t.CRASHScoreMap[userid].GetSum()
				msg.BackMultiple = v
			}
		default:
			tt = t.BetTime - t.timer/20
		}
	}
	if tt < 0 {
		tt = 0
	}
	msg.Timer = uint32(tt)
}

// 进入消息
func (t *Desk) freeCameinMsg(userid string) {
	msg := new(pb.CRASHCameinNtf)
	msg.Userinfo = t.freeSeatRoleMsg(userid)
	t.broadcast4(msg)
}

// 位置上玩家数据
func (t *Desk) freeSeatRoleMsg(userid string) (msg *pb.CRASHFreeUser) {
	var user *data.User
	v := t.roles[userid]
	if v == nil {
		user = t.NewbiewRobot[userid]
	} else {
		user = v.User
	}
	msg = handler.PackCrashFreeUser(user)
	if t.DeskFree == nil {
		return
	}
	// 赢分
	if w, ok := user.FreeWinMap[int32(pb.CRASH)]; ok {
		for _, v2 := range w {
			if v2.Score > 0 {
				msg.Wins += v2.Score
			}
		}
	}
	msg.Bets = t.Bets[userid]
	// if v, ok := t.roles[userid]; ok {
	// 	// if v.Seat == 0 {
	// 	// 	return //没有坐下不广播
	// 	// }
	// 	msg = handler.PackCrashFreeUser(v.User)
	// 	if t.DeskFree == nil {
	// 		return
	// 	}

	// }
	return
}

// 所有坐下玩家数据
func (t *Desk) freeSeatBetsMsg() (msg []*pb.CRASHFreeUser) {
	for _, v := range t.roles {
		msg2 := t.freeSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	for k, u := range t.NewbiewRobot {
		usermsg := handler.PackCrashFreeUser(u)
		if t.DeskFree == nil {
			continue
		}
		// 胜场
		if w, ok := u.FreeWinMap[int32(pb.CRASH)]; ok {
			for _, v2 := range w {
				if v2.Wtype == 1 {
					usermsg.Wins += v2.Score
				}
			}
		}
		usermsg.Bets = t.Bets[k]
		msg = append(msg, usermsg)
	}
	return
}

// 排行榜数据
func (t *Desk) _freeRankMsg(num int32) (msg []*pb.CRASHFreeUser) {
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

// 打包下注排行榜
func (t *Desk) packFreeBetRank() []*pb.CRASHBetRankUser {
	len := min(10, len(t.CRASHDeskFree.BetRank))
	betRank := t.CRASHDeskFree.BetRank[:len]
	for _, rank := range betRank {
		if v, ok := t.DeskFree.CRASHBack[rank.Userid]; ok {
			rank.Multiple0 = v

			if v, ok := t.DeskFree.CRASHScoreMap[rank.Userid]; ok && v.GetSum() > 0 {
				rank.Score += v.GetSum()
			}
		}
		if v, ok := t.DeskFree.CRASHBack1[rank.Userid]; ok {
			rank.Multiple1 = v

			if v, ok := t.DeskFree.CRASHScoreMap1[rank.Userid]; ok && v.GetSum() > 0 {
				rank.Score += v.GetSum()
			}
		}
	}
	return betRank
}

// 更新下注排行榜
func (t *Desk) updateFreeBetRank(userid string, bets int64) {
	var user *data.User
	if v, ok := t.roles[userid]; ok && v != nil {
		user = v.User
	} else {
		user = t.NewbiewRobot[userid]
	}
	if user == nil {
		return
	}

	newRank, curRank := -1, -1
	betRank := t.CRASHDeskFree.BetRank
	if len(betRank) == 0 {
		if bets > 0 {
			newRank = 0
		}
	} else {
		for i, item := range betRank {
			if newRank == -1 && bets >= item.Bets {
				newRank = i
				if curRank >= 0 {
					break
				}
			}
			if curRank == -1 && userid == item.Userid {
				curRank = i
				if newRank >= 0 {
					break
				}
			}
		}
	}
	if curRank >= 0 {
		v := betRank[curRank]
		v.Bets = bets
		// 移除当前的排名
		betRank = utils.SliceRemoveAt(betRank, curRank)
		// 添加新的排名
		if newRank >= 0 {
			betRank = utils.SliceInsertAt(betRank, newRank, v)
		}
		t.CRASHDeskFree.BetRank = betRank
	} else if newRank >= 0 {
		// 插入新的排名
		betRank = utils.SliceInsertAt(betRank, newRank, &pb.CRASHBetRankUser{
			Userid:   user.Userid,
			Nickname: user.Nickname,
			Photo:    user.Photo,
			VipLv:    int32(user.Vip.Lv),
			Bets:     bets,
		})
		// 返10个, 多存几个取消打码时排行榜有值
		if len(betRank) > 16 {
			betRank = betRank[:16]
		}
		t.CRASHDeskFree.BetRank = betRank
	}
}

// // 玩家下注数据
// func (t *Desk) freeBetsMsg(userid string) (msg []*pb.CRASHRoomBets) {
// 	for i := pb.DESK_SEAT2; i <= pb.DESK_SEAT4; i++ {
// 		seat := uint32(i)
// 		bets := t.getFreeSeatBet(userid, seat)
// 		msg2 := &pb.CRASHRoomBets{
// 			Seat: seat,
// 			Bets: bets,
// 		}
// 		msg = append(msg, msg2)
// 	}
// 	return
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
// 				// msg := handler.CRASHBeDealerMsg(0, int64(num), t.DeskFree.CarryInit, t.DeskGame.Dealer,
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
// 	case int32(pb.DEALER_CRASH):
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
// func (t *Desk) dealerListMsg() (msg *pb.CRASHFreeDealerListRsp) {
// 	msg = new(pb.CRASHFreeDealerListRsp)
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
// 		msg2 := &pb.CRASHDealerList{
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

// 百人取消下注
func (t *Desk) freeCancelBet(userid string, nextBet bool, pos int32) (rsp *pb.CRASHCancelBetRsp) {
	rsp = new(pb.CRASHCancelBetRsp)
	_, ok := t.roles[userid]
	if !ok {
		glog.Errorf("userid %s not exist", userid)
		rsp.Error = pb.NotInRoom
		return
	}
	var crashBets map[string]data.Currency
	if nextBet {
		// 取消下一局下注
		crashBets = t.PosCrashNextBetsMap(pos)
	} else {
		// 取消本局下注
		crashBets = t.PosCrashBetsMap(pos)
	}
	bet, ok := crashBets[userid]
	if !ok {
		rsp.Error = pb.NoBet
		return
	}

	t.sendCurrency(userid, bet.Coin, bet.Diamond, int32(pb.LOG_TYPE146), fmt.Sprintf("CRASH房间%s位置%d取消%s下注", t.DeskData.Rid, pos+1, utils.CaseElse(!nextBet, "", "下一轮")))
	delete(crashBets, userid)

	// 可提现金加回去
	if bet.Out > 0 {
		t.checkGiveDiamond(userid, data.Currency{Diamond: bet.Out})
	}

	if !nextBet {
		t.DeskFree.Bets[userid] -= bet.GetSum()
		if t.DeskFree.Bets[userid] <= 0 {
			delete(t.DeskFree.Bets, userid)
		}
		// 更新打码排行榜
		t.updateFreeBetRank(userid, t.DeskFree.Bets[userid])
	}

	players := int32(len(t.DeskFree.Bets))
	var pool int64
	for _, bet := range t.DeskFree.Bets {
		pool += bet
	}
	// 取消打码广播
	ntf := &pb.CRASHCancelBetNtf{
		Userid:  userid,
		Pos:     pos,
		Bets:    bet.GetSum(),
		Pool:    pool,
		NextBet: nextBet,
		Players: players,
	}
	t.broadcast2(userid, ntf)

	rsp.Userid = userid
	rsp.Pos = pos
	rsp.Bets = bet.GetSum()
	rsp.Pool = pool
	rsp.NextBet = nextBet
	rsp.Players = players
	return
}

// '百人下注
func (t *Desk) freeBet(userid string, num int64, pos int32) pb.ErrCode {
	// 不在下注状态, 记到下一局下注
	// if t.state != int32(pb.STATE_BET) {
	// 	return pb.BetOver
	// }
	if num <= 0 {
		return pb.OperateError
	}
	if table.GetTables().CrashRoomTable.Get().Status == 0 {
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
	stateInBet := t.state == int32(pb.STATE_BET)

	var crashBets map[string]data.Currency
	if stateInBet {
		crashBets = t.PosCrashBetsMap(pos)
	} else {
		crashBets = t.PosCrashNextBetsMap(pos)
	}

	// 下注限制
	betLimit := table.GetTables().CrashTable.Get(pos + 1).BetLimit[user.RegistArea]
	if c, ok := crashBets[userid]; ok {
		if c.GetSum()+num > int64(betLimit.Value[1]) {
			return pb.BetTopLimit //tie下注限制
		}
	} else if num < int64(betLimit.Value[0]) || num > int64(betLimit.Value[1]) {
		return pb.BetTopLimit //tie下注限制
	}
	cashBet := t.getCashOrCoinBet(role, num, true)  // 彩金下注
	coinBet := t.getCashOrCoinBet(role, num, false) // 奖励金下注
	t.sendCurrency(userid, (-1 * (num - cashBet)), (-1 * cashBet), int32(pb.LOG_TYPE79), fmt.Sprintf("CRASH房间%s位置%d%s下注", t.DeskData.Rid, pos+1, utils.CaseElse(stateInBet, "", "下一轮")))

	//下注记录
	if stateInBet {
		t.DeskFree.Bets[userid] += num //个人总下注额
	}

	// if stateInBet && !user.GetRobot() {
	// 	t.DeskGame.CashBets += cashBet //当局彩金总下注额
	// 	t.DeskGame.CoinBets += coinBet //当局奖励金总下注额
	// 	if user.OutDiamond-user.Diamond > 0 {
	// 		t.BackOutDiamond[userid] = user.OutDiamond - user.Diamond
	// 	}
	// }

	//位置详细记录
	bet := data.Currency{Diamond: cashBet, Coin: coinBet}
	if user.OutDiamond > user.Diamond {
		// 记录可提现金, 返奖/取消下注时退还
		bet.Out = min(bet.Diamond, user.OutDiamond-user.Diamond)
		t.checkOutDiamond(userid)
	}
	if m, ok := crashBets[userid]; ok {
		m.Diamond += bet.Diamond
		m.Coin += bet.Coin
		m.Out += bet.Out
		crashBets[userid] = m
	} else {
		crashBets[userid] = bet
	}

	if stateInBet {
		// 更新打码排行榜
		t.updateFreeBetRank(userid, t.DeskFree.Bets[userid])
	}

	role.NoBetTimes = 0 // 不下注次数清零
	msg, ntf := resFreeBet(cashBet+coinBet, crashBets[userid].GetSum(), t.DeskFree.Bets, user, pos, stateInBet)
	t.send2userid(userid, msg)
	t.broadcast2(userid, ntf)
	return pb.OK
}

// '新手百人下注
func (t *Desk) newBiewFreeBet(userid string, num int64, pos int32) pb.ErrCode {
	if t.state != int32(pb.STATE_BET) {
		return pb.BetOver
	}
	status := table.GetTables().CrashRoomTable.Get().Status
	if status == 0 {
		return pb.RoomMaintenance
	}
	user := t.NewbiewRobot[userid]
	if user == nil {
		glog.Errorf("newbiew robot userid %s not exist", userid)
		return pb.NotInRoom
	}
	betLimit := table.GetTables().CrashTable.Get(pos + 1).BetLimit[0]
	crashBets := t.PosCrashBetsMap(pos)
	if c, ok := crashBets[userid]; ok {
		if c.GetSum()+num > int64(betLimit.Value[1]) {
			return pb.BetTopLimit //tie下注限制
		}
	}
	user.AddCoin(-num)
	t.DeskFree.Bets[userid] += num //房间总下注额
	//位置详细记录
	if m, ok := crashBets[userid]; ok {
		m.Coin += num
		crashBets[userid] = m
	} else {
		m := data.Currency{Coin: num}
		crashBets[userid] = m
	}

	// 更新打码排行榜
	t.updateFreeBetRank(userid, t.DeskFree.Bets[userid])

	_, ntf := resFreeBet(num, crashBets[userid].GetSum(), t.DeskFree.Bets, user, pos, true)
	t.broadcast(ntf)
	return pb.OK
}

// 将下一轮下注提出来下注
func (t *Desk) nextBetBetting() {
	for _, pos := range []int32{0, 1} {
		nextBets := t.PosCrashNextBetsMap(pos)
		crashBets := t.PosCrashBetsMap(pos)

		for userid, bet := range nextBets {
			role, ok := t.roles[userid]
			if !ok || role == nil {
				continue
			}

			bet.SetTag(TagNextBet)
			crashBets[userid] = bet

			t.DeskFree.Bets[userid] += bet.GetSum() //个人总下注额
			role.NoBetTimes = 0                     //不下注次数清零

			// 更新打码排行榜
			t.updateFreeBetRank(userid, t.DeskFree.Bets[userid])
			msg, ntf := resFreeBet(bet.GetSum(), crashBets[userid].GetSum(), t.DeskFree.Bets, role.User, pos, true)
			ntf1 := &pb.CRASHNextBetNtf{Rsp: msg}
			t.send2userid(userid, ntf1)
			t.broadcast2(userid, ntf)
		}
	}
	t.CrashNextBets = make(map[string]data.Currency)
	t.CrashNextBets1 = make(map[string]data.Currency)
}

// 下注消息
func resFreeBet(val, posBets int64, roomBets map[string]int64, user *data.User, pos int32, stateInBet bool) (*pb.CRASHBetRsp, *pb.CRASHBetNtf) {
	var pool, bets int64
	for k, bet := range roomBets {
		pool += bet
		if k == user.Userid {
			bets = bet
		}
	}
	userid := user.Userid
	payers := int32(len(roomBets))
	rsp := &pb.CRASHBetRsp{
		Value:   uint32(val),
		Pool:    pool,
		Bets:    bets,
		PosBets: posBets,
		Userid:  userid,
		Pos:     pos,
		NextBet: !stateInBet,
		Players: payers,
	}
	ntf := &pb.CRASHBetNtf{
		Value:    uint32(val),
		Pool:     pool,
		Bets:     bets,
		PosBets:  posBets,
		Userid:   userid,
		Pos:      pos,
		NextBet:  !stateInBet,
		Nickname: user.Nickname,
		Photo:    user.Photo,
		VipLv:    int32(user.Vip.Lv),
		Players:  payers,
	}
	return rsp, ntf
}

//.

//'游戏状态

//.

//.

// '结束游戏
func (t *Desk) freeGameOver() {
	defer func() {
		t.tickStop = false
	}()
	//赔付
	t.gameSettlement()
	//对局详情, 有玩家再保存
	if hasPlayer := len(t.roles) > 0; hasPlayer {
		t.gameDetail()
	}
	//打印信息
	// t.printOver()
	//结束消息广播
	msg := t.resOverFree()
	t.broadcast4(msg)
	// for k, role := range t.roles {
	// 	if role.Offline || role.Robot {
	// 		continue
	// 	}
	// 	// msg.WinningScore = t.CRASHPlayerJackpot[k]
	// 	t.send2userid(k, msg)
	// }

	// 策略结算
	// for k := range t.CRASHScoreMap {
	// 	t.strategySettlement(k)
	// }

	// t.broadcast4(msg)
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
	//消息广播
	// t.freeStart()
	//奖池清零
	t.DeskGame.CashBets = 0
	t.DeskGame.CoinBets = 0
	//赠送金库存变化
	if t.GiveStock > 0 {
		// t.changeStock(t.GiveStock, 0, 0, 0, 0, 0, "50033")
	}
}

// .结账
func (t *Desk) gameSettlement() {
	glog.Infof("rocket boom ,rid:%s,gameid:%s,stockid:%s,Mulpitle:%.2f", t.Rid, t.GameId, "", float32(t.CRASHMulpitle)/100)

	crashScoreMap := make(map[string]data.Currency)
	// 已撤离的
	for _, pos := range []int32{0, 1} {
		scoreMap := t.PosCrashScoreMap(pos)
		for userid, score := range scoreMap {
			c := crashScoreMap[userid]
			c.Merge(score)
			crashScoreMap[userid] = c
		}
	}

	// 没撤离的
	for _, pos := range []int32{0, 1} {
		posBets := t.PosCrashBetsMap(pos)
		posBack := t.PosCrashBackMap(pos)
		scoreMap := t.PosCrashScoreMap(pos)
		for k, bets := range posBets {
			if _, ok := posBack[k]; ok {
				// 已撤离
				continue
			}

			lose := bets.Scale(-1)
			// 记录输赢分
			scoreMap[k] = lose
			// 赠送金变化
			t.WinScoreSettlement(k, lose)

			c := crashScoreMap[k]
			c.Merge(lose)
			crashScoreMap[k] = c
		}
	}

	for k, c := range crashScoreMap {
		score := c.GetSum()
		var role *data.User
		if v, ok := t.roles[k]; ok && v != nil {
			role = v.User
		} else {
			// 人机
			role = t.NewbiewRobot[k]
		}
		if role == nil {
			continue
		}

		// 打码量
		var bets int64
		if v, ok := t.DeskFree.Bets[k]; ok {
			bets = v
		}
		// 明税
		var mcash int64
		if v, ok := t.CRASHTaxMap[k]; ok {
			mcash = v
		}
		if !role.Robot {
			// 对局详情
			t.crashDetail(role, score, mcash, 0, 0, 0)
		} else if score > 0 && t.getBackBigMultiple(role.Userid) > 1000 {
			// 人机逃脱了倍数大于10才记录(crash大赢家榜)
			t.crashDetail(role, score, mcash, 0, 0, 0)
		}

		if !role.GetRobot() {
			// 事件(非机器人)
			bean := &event.GameRecordEvent{Gtype: uint32(pb.CRASH), Win: score > 0}
			t.eventPost(k, event.PLAY_TASK, bean) // 玩游戏
			// 分享打码量
			t.shareAmount(k, bets)
			// 游戏时长
			t.GameTime(k)

			// 逃离倍数
			var multiple, multiple1 int32
			if v, ok := t.DeskFree.CRASHBack[k]; ok {
				multiple = v
			}
			if v, ok := t.DeskFree.CRASHBack1[k]; ok {
				multiple1 = v
			}
			// vip bank 游戏任务
			t.eventPost(k, event.VB_GAME_TASK, &event.VBGameTaskEvent{
				GameType:     int32(pb.CRASH),
				Win:          score > 0,
				CrashEscape:  float64(multiple) / 100.0,
				CrashEscape1: float64(multiple1) / 100.0,
			})
		}
		// 玩家输赢记录
		t.CRASHRecord(k, score)
		// t.strategySettlement(k)

		// 事件
		// t.eventPost(k, event.Crash_Settlement, &event.CrashSettlementEvent{Score: score, Multiple: multiple, Multiple1: multiple1, Bet: bets})
	}

	for _, v := range t.CRASHObserver {
		bet := t.DeskFree.Bets[v]
		if role, ok := t.roles[v]; ok && !role.Robot && bet <= 0 {
			t.crashDetail(role.User, 0, 0, 0, 0, 0)
		}
	}

	// 游戏开奖记录
	t.addHistory()
	t.CRASHDetail.DeskId = "" // t.Game.Id
	t.CRASHDetail.CRASHDetail.Result = fmt.Sprintf("x%.2f", float32(t.CRASHMulpitle)/100)
}

// 获取最大逃脱倍数
func (t *Desk) getBackBigMultiple(userid string) int32 {
	var multiple, multiple1 int32
	if v, ok := t.CRASHDeskFree.CRASHBack[userid]; ok {
		multiple = v
	}
	if v, ok := t.CRASHDeskFree.CRASHBack1[userid]; ok {
		multiple1 = v
	}
	return max(multiple, multiple1)
}

// 输赢记录
func (t *Desk) CRASHRecord(userid string, num int64) {
	bets := t.DeskFree.Bets[userid]
	msg := &pb.FreeSetRecord{Gtype: int32(pb.CRASH), Score: num, Bet: bets}
	var role *data.User
	var rolePid *actor.PID
	if v, ok := t.roles[userid]; ok && v != nil {
		role = v.User
		rolePid = v.Pid
	} else if v, ok := t.NewbiewRobot[userid]; ok && v != nil {
		role = v
	} else {
		return
	}

	var result int32
	if num > 0 {
		// 赢了
		result = 1
	} else if num == 0 {
		result = 0
	} else {
		result = -1
	}
	msg.Rtype = result
	msg.GameTime = utils.BsonNow().Unix() - t.DeskGame.BeginTime
	t.send2userid(userid, msg)

	// 修改玩家数据
	wins := make([]data.FreeWin, 0)
	if role.FreeWinMap == nil {
		role.FreeWinMap = make(map[int32][]data.FreeWin)
	}
	if w, ok := role.FreeWinMap[int32(pb.CRASH)]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: result, Score: num})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}
	role.FreeWinMap[int32(pb.CRASH)] = wins
	role.BetRecord = append(role.BetRecord, bets)
	if len(role.BetRecord) > 200 {
		role.BetRecord = role.BetRecord[len(role.BetRecord)-200:]
	}

	if role.RoundGames == nil {
		role.RoundGames = make(map[int32]int32)
	}
	role.RoundGames[int32(pb.CRASH)]++
	if role.RoundGames[int32(pb.CRASH)] <= 1 {
		t.CRASHFirstPartIn = true
	}

	// 打码量日志
	if !role.Robot {
		// 打码上报
		if err = mq.NatsPublish(mq.TopicGameBets, &pb.PublishGameBets{
			UserPid:    rolePid,
			Userid:     role.Userid,
			Gtype:      int32(pb.CRASH),
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
			glog.Error("publish user crash bets error", err)
		}

		log := handler.GameFlowWaterLog(bets, int(pb.CRASH), 0, len(role.RechargeTarge), userid)
		myactor.Logger().Tell(log)
	}
}

// 开奖记录
func (t *Desk) addHistory() {
	t.DeskFree.CRASHHistory = append(t.DeskFree.CRASHHistory, t.CRASHMulpitle)
	size := len(t.DeskFree.CRASHHistory)
	glog.Infof("history size:", size)
	if size > 50 { // 只保留50条
		t.DeskFree.CRASHHistory = t.DeskFree.CRASHHistory[size-50:]
	}
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
func (t *Desk) resOverFree() *pb.CRASHGameoverNtf {
	msg := &pb.CRASHGameoverNtf{
		State:    t.state,
		Userinfo: t.freeSeatBetsMsg(),
		Multiple: fmt.Sprintf("%.2f", float32(t.CRASHMulpitle)/100),
		Jackpot:  t.CrashJackpot,
	}
	crashScoreMap := make(map[string]int64)
	for userid, c := range t.CRASHScoreMap {
		crashScoreMap[userid] += c.GetSum()
	}
	for userid, c := range t.CRASHScoreMap1 {
		crashScoreMap[userid] += c.GetSum()
	}
	for k, v := range crashScoreMap {
		bean := &pb.CRASHRoomScore{
			Userid: k,
			Score:  v,
		}
		msg.Scores = append(msg.Scores, bean)
	}
	return msg
}

// 赢分结算
func (t *Desk) WinScoreSettlement(userid string, score data.Currency) {
	// 跑马灯检测
	t.checkMarquee(userid, score)
	// 赠送金变化
	t.checkGiveDiamond(userid, score)
	// 罐子流水
	t.flowWater(userid, score.Diamond)
	// 点控房间
	// if t.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
	// 	t._controlSettlement(userid, score.Diamond)
	// }
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
		if t.timer >= 20 {
			t.timer = 0
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_BET)
			msg.During = int32(t.BetTime)
			t.selfPid.Tell(msg)
			return
		}
	// case int32(pb.STATE_DEALING):
	// 	// 发牌状态
	// 	if t.timer >= FreeDealTime {
	// 		t.timer = 0
	// 		t.state = int32(pb.STATE_BET) // 下注状态
	// 		msg := new(pb.ChangeFreeDeskStatus)
	// 		msg.Status = int32(pb.STATE_BET)
	// 		msg.During = FreeBetTime
	// 		t.selfPid.Tell(msg)
	// 		return
	// 	}
	case int32(pb.STATE_BET):
		// 下注状态
		if t.timer >= t.BetTime*20 {
			t.timer = 0
			t.state = int32(pb.STATE_PAUSE) // 等待状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_PAUSE)
			msg.During = 1
			t.selfPid.Tell(msg)
			return
		}
	case int32(pb.STATE_PAUSE):
		// 下注状态
		if t.timer >= 20 {
			t.timer = 0
			// t.state = int32(pb.STATE_PAUSE) // 开牌状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_LEAD)
			msg.During = 1
			t.selfPid.Tell(msg)
			t.tickStop = true
			t.CRASHDeskFree.CrashBoom = false
			return
		}
	case int32(pb.STATE_LEAD):
		// 开牌状态
		if t.CRASHLeadTime > 0 && t.timer >= t.CRASHLeadTime {
			t.timer = 0
			t.state = int32(pb.STATE_OVER) // 结算状态
			msg := new(pb.ChangeFreeDeskStatus)
			msg.Status = int32(pb.STATE_OVER)
			msg.During = FreeSettlementTime
			t.selfPid.Tell(msg)
			t.tickStop = true
			return
		}
	case int32(pb.STATE_OVER):
		// 结算状态
		if t.timer >= FreeSettlementTime*20 {
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

func (t *Desk) boomCheck() {
	if t.state != int32(pb.STATE_LEAD) || t.CRASHDeskFree.CrashBoom {
		return
	}

	msg := new(pb.CrashBoom)
	t.selfPid.Tell(msg)
	t.CRASHDeskFree.CrashBoom = true
}

// 火箭起飞
func (t *Desk) takeoff() {
	// 计算爆炸倍数爆炸时间
	t.CRASHRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	t.CRASHBoomMulpitle = handler.CrashBoomMultiple(t.CRASHRand)
	now := time.Now()
	t.CRASHTakeoffTime = now
	boomDuration := handler.CrashBoomDuration(t.CRASHBoomMulpitle)
	t.CRASHBoomTime = now.UnixMilli() + boomDuration.Milliseconds()
	// glog.Infof("--------------takeoff: %v, %v", t.CRASHBoomMulpitle, t.CRASHBoomTime, boomDuration.Milliseconds())

	t.state = int32(pb.STATE_LEAD)
}

// 火箭起飞
// Deprecated
func (t *Desk) _takeoff() {
	room, err := t._getRoom()
	if err != nil {
		// 异常牌局
		return
	}
	t.Game = room
	// var factor float64 = 0
	// var stock int64 = 0
	// switch t.Dtype {
	// case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B):
	// 	factor, stock = t.getFinalFactor()
	// case int32(pb.DESK_TYPE_POINTCONTROL):
	// 	factor, stock = t.getControlFactor()
	// }
	// if t.DeskGame.CashBets+t.DeskGame.CashBets <= 0 {
	// 	// 没人下注,随机一个倍数
	// 	t.CRASHControlMulpitle = t.randomMulpitle()
	// }
	// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) && t.CashBets > 0 {
	// 	// 新手房间并且有人下注
	// 	t.CRASHControlMulpitle = t.getNewbiewMulpitle()
	// }

	// B类玩家触发必赢策略
	// if m, ok := t.WinStrategy_B(); ok {
	// 	t.CRASHControlMulpitle = m
	// }
	// if env == "dev" {
	// 	// test环境
	// 	cfg.Reload()
	// 	boomMulpitle := cfg.Section(nodeName).Key("boom").Value()
	// 	if boomMulpitle != "" {
	// 		b, err := strconv.Atoi(boomMulpitle)
	// 		if err == nil {
	// 			t.CRASHControlMulpitle = int32(b)
	// 		}
	// 	}
	// }

	// 选择策略
	t.selectStrategy()

	// 初始化随机种子
	r = rand.New(rand.NewSource(time.Now().UnixNano()))

	// t.DeskFree.CRASHRealStock = stock
	// t.DeskFree.CRASHFactor = factor
	// t.DeskFree.CRASHDetail.PlayerFactor = int(factor * 100)
	t.state = int32(pb.STATE_LEAD)
	// glog.Infof("crash factor %.2f, stock:%s, gameId:%s, maxMulpitle:%d", factor, room.Id, t.GameId, t.CRASHControlMulpitle)
}

// 根据下注总额获取对应的库存
func (t *Desk) _getRoom() (data.Game, error) {
	// 选择走哪个库存
	pool := t.DeskGame.CashBets
	games := config.GetGames()
	var roomId int32 = 0
	for _, g := range games {
		if g.Gtype != int32(pb.CRASH) || g.Dtype != t.Dtype {
			continue
		}
		switch t.Dtype {
		case int32(pb.DESK_TYPE_POINTCONTROL): // 点控
			roomId = g.CRASH.ControlRoom[0]
		// case int32(pb.DESK_TYPE_NEWBIEW): // 新手
		// 	roomId = g.CRASH.NoviceRoom[0]
		case int32(pb.DESK_TYPE_NORMAL), int32(pb.DESK_TYPE_NORMAL_B),
			int32(pb.DESK_TYPE_NEWBIEW): // 正常
			intervals := g.CRASH.BetInterval
			for i, v := range intervals {
				if pool >= int64(v[0]) && v[1] == -1 || (pool >= int64(v[0]) && pool < int64(v[1])) {
					roomId = g.CRASH.BetRoom[i][0]
					break
				}
			}
		}
	}
	room := config.GetGame(utils.String(roomId))
	if room.Id == "" {
		return room, fmt.Errorf("no find room, id:%d", roomId)
	}
	return room, nil
}

// 更新排行榜
func (t *Desk) _CRASHdateRank(userid string) {
	rank := t._freeRankMsg(20)
	if userid != "" {
		CRASHdate := false
		for _, lu := range rank {
			if lu.Userid == userid {
				CRASHdate = true
			}
		}
		if CRASHdate {
			msg := new(pb.CRASHRankNtf)
			msg.Rank = rank
			t.broadcast(msg)
		}
		return
	}
	msg := new(pb.CRASHRankNtf)
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
		glog.Errorf("CRASH event Marshal fail")
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
func (t *Desk) gameDetail() {
	if len(t.CRASHDetail.CRASHDetail.UserDetail) <= 0 {
		return
	}
	// 保存详情
	t.CRASHDeskFree.CRASHDetail.EndTime = utils.BsonNow().Unix()
	players := make([]string, 0)
	for _, ld := range t.CRASHDetail.CRASHDetail.UserDetail {
		if ld.Observe || ld.Robot {
			continue
		}
		players = append(players, ld.Userid)
	}
	t.CRASHDetail.Players = strings.Join(players, ",")
	for _, s := range t.CRASHTriggerStrategys {
		t.CRASHDetail.CRASHDetail.StrategyType = append(t.CRASHDetail.CRASHDetail.StrategyType, s.Id)
	}
	body, err := json.Marshal(t.CRASHDeskFree.CRASHDetail)
	if err != nil {
		glog.Errorf("save detail error %#v", t.CRASHDeskFree.CRASHDetail)
		return
	}
	msg := &pb.Detail{Data: body}
	myactor.Logger().Tell(msg)
}

// 更新排行榜
func (t *Desk) _updateRank(userid string) {
	rank := t._freeRankMsg(20)
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
			msg := new(pb.CRASHRankNtf)
			msg.Rank = rank
			t.broadcast4(msg)
		}
		return
	}
	msg := new(pb.CRASHRankNtf)
	msg.Rank = rank
	t.broadcast4(msg)
}

// 爆炸点
func (t *Desk) boomPoint() {
	now := time.Now()
	multiple := handler.CrashTimeMultiple(t.CRASHTakeoffTime, now) // 当前倍数
	multiple = min(multiple, t.DeskFree.CRASHBoomMulpitle-1)
	t.CRASHMulpitle = multiple // 开奖倍数增加

	// !now.Before(t.CRASHBoomTime) ||
	if now.UnixMilli() >= t.CRASHBoomTime || multiple >= t.DeskFree.CRASHBoomMulpitle {
		glog.Infof("crash boomPoint: %v, %v, %v, %v", t.CRASHBoomTime, now.UnixMilli(), multiple, t.DeskFree.CRASHBoomMulpitle)

		//爆炸了
		t.CRASHMulpitle = t.DeskFree.CRASHBoomMulpitle - 1

		//先检查自动逃离的
		t.CRASHDeskFree.CrashBoom = false
		t.checkAutoBack(true)

		rsp := new(pb.CRASHBoomNtf)
		rsp.Multiple = multiple
		t.broadcast4(rsp)
		//切换状态
		t.CRASHLeadTime = 40 // 爆炸后等待时长 40*50=2s
		t.CRASHDeskFree.CrashBoom = true
		t.tickStop = false
		return
	}

	//先检查自动逃离的
	t.CRASHDeskFree.CrashBoom = false
	t.checkAutoBack(false)
	// glog.Debugf("Mulpitle %.2f", t.CRASHMulpitle)

	//通知客户端
	// crashBets := t.PosCrashBetsMap(0)
	// crashBets1 := t.PosCrashBetsMap(1)
	// for userid, v := range t.roles {
	// 	if v.Offline || v.Robot {
	// 		continue
	// 	}
	// 	ntf := &pb.CRASHMultipleNtf{Multiple: t.CRASHMulpitle}
	// 	// 当前返奖额
	// 	if c, ok := crashBets[userid]; ok {
	// 		win := c.Scale(float64(multiple) / 100)
	// 		ntf.Win = win.GetSum()
	// 	}
	// 	if c, ok := crashBets1[userid]; ok {
	// 		win := c.Scale(float64(multiple) / 100)
	// 		ntf.Win1 = win.GetSum()
	// 	}
	// 	v.Pid.Tell(ntf)
	// }
}

// 撤离
func (t *Desk) back(userid string, auto bool, multiple, pos int32) {
	rsp := new(pb.CRASHBackRsp)
	if t.state != int32(pb.STATE_LEAD) || t.CRASHDeskFree.CrashBoom {
		// 撤离失败，火箭已经爆炸了
		if auto {
			glog.Infof("crash auto leave error: %s, %d, %d, %d", userid, pos, multiple, t.CRASHDeskFree.CRASHBoomMulpitle)
		}
		rsp.Error = pb.CrashFail
		t.send2userid(userid, rsp)
		return
	}
	if multiple <= 0 {
		multiple = handler.CrashTimeMultiple(t.DeskFree.CRASHTakeoffTime, time.Now())
	}
	if multiple >= t.CRASHDeskFree.CRASHBoomMulpitle {
		// 撤离失败，火箭已经爆炸了
		if auto {
			glog.Infof("crash auto leave error: %s, %d, %d, %d", userid, pos, multiple, t.CRASHDeskFree.CRASHBoomMulpitle)
		}
		rsp.Error = pb.CrashFail
		t.send2userid(userid, rsp)
		return
	}

	role := t.roles[userid]
	if role == nil {
		// 人机撤离
		t.robotBack(userid, multiple, pos)
		return
	}

	// 下注
	crashBets := t.PosCrashBetsMap(pos)
	bets, ok := crashBets[userid]
	if !ok {
		rsp.Error = pb.NoBet
		t.send2userid(userid, rsp)
		return
	}

	win := bets.Scale(float64(multiple) / 100)
	// 返奖上限
	winLimit := table.GetTables().CrashTable.Get(pos + 1).WinLimit[0]
	if win.GetSum() > winLimit {
		win = data.Currency{Diamond: winLimit}
	}
	win.Sub(bets) // 净输赢分

	//明税
	var mcash int64
	if win.GetSum() > 0 {
		taxRate := table.GetTables().CrashRoomTable.Get().TaxRate[role.RegistArea]
		if taxRate > 0 {
			mcash = int64(math.Ceil(float64(win.Diamond) * float64(taxRate) / 10000))
			win.Sub(data.Currency{Diamond: mcash})
		}
	}
	t.sendCurrency(userid, bets.Coin+win.Coin, bets.Diamond+win.Diamond, int32(pb.LOG_TYPE73), fmt.Sprintf("CRASH房间%s位置%d撤离", t.DeskData.Rid, pos+1))
	if bets.Out > 0 {
		t.checkGiveDiamond(userid, data.Currency{Diamond: bets.Out})
	}

	if auto {
		win.SetTag(TagAutoCrash) // 记录自动逃离
	}
	// 逃脱倍数
	crashBack := t.PosCrashBackMap(pos)
	crashBack[userid] = multiple
	// 玩家输赢分
	crashScoreMap := t.PosCrashScoreMap(pos)
	crashScoreMap[userid] = win
	// 抽水
	if mcash > 0 {
		t.CRASHTaxMap[userid] = mcash
	}
	// 已撤离分数
	t.CRASHLose += win.Diamond + bets.Diamond

	if !role.GetRobot() {
		glog.Infof("[crash] user %s back, win:%d, multiple:%d", userid, win.GetSum(), multiple)
	}

	// 赠送金变化
	t.WinScoreSettlement(userid, win)

	if auto {
		//自动逃离
		a := &pb.CRASHAutoBackNtf{
			Multiple: multiple,
			Score:    win.GetSum() + bets.GetSum(),
			Jackpot:  0,
			Pos:      pos,
		}
		t.send2userid(userid, a)
	} else {
		rsp.Multiple = multiple
		rsp.Score = win.GetSum() + bets.GetSum()
		rsp.Jackpot = 0
		rsp.Pos = pos
		t.send2userid(userid, rsp)
	}

	// 通知其他玩家
	ntf := &pb.CRASHBackNtf{
		Name:     role.Nickname,
		Multiple: multiple,
		Pos:      pos,
		Userid:   role.Userid,
		Photo:    role.Photo,
		VipLv:    int32(role.Vip.Lv),
		Bets:     bets.GetSum(),
		Score:    win.GetSum() + bets.GetSum(),
	}
	t.broadcast4(ntf)
}

// 玩家对局详情
func (t *Desk) crashDetail(role *data.User, score, mcash, mcoin, acash, acoin int64) {
	userid := role.Userid
	// role := t.roles[userid]
	bet := t.DeskFree.Bets[userid]
	result := t.getWinReslut(score)
	// 本局打码
	var bet0, bet1 data.Currency
	if v, ok := t.CRASHDeskFree.CrashBets[userid]; ok {
		bet0 = v
	}
	if v, ok := t.CRASHDeskFree.CrashBets1[userid]; ok {
		bet1 = v
	}
	// 逃脱倍数
	var multiple, multiple1 string
	if v, ok := t.CRASHDeskFree.CRASHBack[userid]; ok {
		multiple = fmt.Sprintf("%.2f", float64(v)/100)
	}
	if v, ok := t.CRASHDeskFree.CRASHBack1[userid]; ok {
		multiple1 = fmt.Sprintf("%.2f", float64(v)/100)
	}
	// 输赢分
	var score0, score1 data.Currency
	if v, ok := t.CRASHDeskFree.CRASHScoreMap[userid]; ok {
		score0 = v
	}
	if v, ok := t.CRASHDeskFree.CRASHScoreMap1[userid]; ok {
		score1 = v
	}

	var autoCrashM0, autoCrashM1 int32
	if v, ok := t.DeskFree.CrashAutoLeave[userid]; ok {
		autoCrashM0 = v
	}
	if v, ok := t.DeskFree.CrashAutoLeave1[userid]; ok {
		autoCrashM1 = v
	}

	// 对局记录
	d := data.CRASHUserDetail{
		Userid:       userid,
		Nickname:     role.Nickname,
		Photo:        role.Photo,
		VipLv:        int32(role.Vip.Lv),
		Bet:          bet,
		Bet0:         bet0.GetSum(),
		Bet1:         bet1.GetSum(),
		Observe:      bet <= 0,
		Multiple:     multiple,
		Multiple1:    multiple1,
		Result:       result,
		Win:          score,
		Win0:         score0.GetSum(),
		Win1:         score1.GetSum(),
		AfterScore:   role.GetScore(),
		AfterCash:    role.Diamond,
		AfterBonus:   role.Coin,
		CashMingTax:  mcash,
		BonusMingTax: mcoin,
		CashAnTax:    acash,
		BonusAnTax:   acoin,
		Robot:        role.Robot,
		NextBet0:     bet0.HasTag(TagNextBet),
		NextBet1:     bet1.HasTag(TagNextBet),
		AutoCrash0:   score0.HasTag(TagAutoCrash),
		AutoCrash1:   score1.HasTag(TagAutoCrash),
		AutoCrashM0:  autoCrashM0,
		AutoCrashM1:  autoCrashM1,
	}

	if score > 0 {
		d.BeforeScore = role.GetScore() - score
		if !role.Robot {
			d.BeforeCash = role.Diamond - score
		} else {
			d.BeforeBonus = role.Coin - score
		}
	} else {
		d.BeforeScore = role.GetScore() + bet
		if !role.Robot {
			d.BeforeCash = role.Diamond + bet
		} else {
			d.BeforeBonus = role.Coin + bet
		}
	}
	t.CRASHDetail.CRASHDetail.Bets += bet
	t.CRASHDetail.CRASHDetail.PlayerLose += score
	t.CRASHDetail.CRASHDetail.UserDetail = append(t.CRASHDetail.CRASHDetail.UserDetail, d)
}

// 点控房间结算
func (t *Desk) _controlSettlement(userid string, cash int64) {
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

// 检查自动逃离
func (t *Desk) checkAutoBack(boom bool) {
	// 位置1自动逃离
	for userid, multiple := range t.DeskFree.CrashAutoLeave {
		// 未下注
		if bets, ok := t.DeskFree.CrashBets[userid]; !ok || bets.GetSum() <= 0 {
			continue
		}
		// 已撤离
		if _, ok := t.DeskFree.CRASHBack[userid]; ok {
			continue
		}
		if (!boom && multiple <= t.CRASHMulpitle) ||
			(boom && multiple < t.CRASHBoomMulpitle) {
			t.back(userid, true, multiple, 0)
		}
	}

	// 位置2自动逃离
	for userid, multiple := range t.DeskFree.CrashAutoLeave1 {
		if bets, ok := t.DeskFree.CrashBets1[userid]; !ok || bets.GetSum() <= 0 {
			continue
		}
		if _, ok := t.DeskFree.CRASHBack1[userid]; ok {
			continue
		}
		if (!boom && multiple <= t.CRASHMulpitle) ||
			(boom && multiple < t.CRASHBoomMulpitle) {
			t.back(userid, true, multiple, 1)
		}
	}
}

// 设置自动逃离
func (t *Desk) setAutoLeave(userid string, autoLeave bool, multiple, pos int32) {
	if !autoLeave {
		switch pos {
		case 0:
			delete(t.DeskFree.CrashAutoLeave, userid)
		case 1:
			delete(t.DeskFree.CrashAutoLeave1, userid)
		}
	} else {
		switch pos {
		case 0:
			t.DeskFree.CrashAutoLeave[userid] = multiple
		case 1:
			t.DeskFree.CrashAutoLeave1[userid] = multiple
		}
	}
}

// 获取新手房间倍数
func (t *Desk) _getNewbiewMulpitle() int32 {
	// bean := table.GetTables().NewbieTable.Get()
	max := t.Game.CRASH.MaxWithdrawable[0] // A类可提现金额上限
	userid := t.GetOnlyOnePlayer()
	if userid != "" {
		role := t.roles[userid]
		if role.IsB() {
			// B类
			max = t.Game.CRASH.MaxWithdrawable[1]
		}
	}
	bet := data.Currency{}
	for k, r := range t.roles {
		bet = t.CrashBets[k]
		if bet.GetSum() <= 0 {
			continue
		}
		canget := int64(max) - r.OutDiamond // 可获得的最大彩金
		maxMutiple := canget*100/bet.Diamond - 1
		return int32(math.Max(float64(maxMutiple), 100))
	}
	return utils.RandInt32N(4901) + 100
}

func (t *Desk) GameTime(userid string) {
	if role, ok := t.roles[userid]; ok {
		if role.Robot {
			return
		}

		msg := &pb.LogGameTime{
			Userid: userid,
			Gtype:  int32(pb.CRASH),
			Time:   time.Now().Unix() - t.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2userid(userid, msg)
	}
}

func (t *Desk) WinStrategy_B() (int32, bool) {
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
			if bet >= 2000 {
				// 玩家只要下注超过20就在9-11倍爆炸
				t.send2userid(role.Userid, &pb.TriggerFreeWelfare{Userid: role.Userid})
				return utils.RandInt32N(201) + 900, true
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

// 获取唯一下注玩家
func (t *Desk) GetOnlyOneBetPlayer() string {
	userids := []string{}
	for k, v := range t.roles {
		if _, ok := t.Bets[k]; ok && !v.User.GetRobot() {
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
