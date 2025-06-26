package andarbahar

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"strconv"
	"time"
)

// '踢出不足坐下玩家或超额玩家
func (t *Desk) limitOver() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
		if !t.DeskData.Pub {
			//return
		}
	case int32(pb.ROOM_TYPE2): //百人
	}
	for k, v := range t.roles {
		ntf := &pb.ABLeaveNtf{Userid: k, Seat: v.Seat}

		var err pb.ErrCode
		if !t.checkGameTime(v) {
			err = pb.KickGameTime
		} else if !t.checkCloseServer() {
			err = pb.KickCloseServer
		} else if !t.check5RoundNoBet(v) {
			err = pb.NoBet5Round
		}
		if err == pb.OK {
			continue
		}

		errcode := t.leave(k, int32(err))
		if errcode != pb.OK {
			continue
		}

		ntf.Error = err
		t.send2userid(k, ntf)

		//离开状态消息
		t.userLeaveDesk(k)
	}
}

// 剔除不满足条件的机器人
func (t *Desk) kickRobot() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		robot := t.Game.AB.Robot // 人机策略 TODO
		for k, v := range t.roles {
			if !v.GetRobot() {
				continue
			}
			score := v.GetScore() // 分数
			if score >= int64(robot.LeaveLimit[0]) && score <= int64(robot.LeaveLimit[1]) {
				// 分数区间正常
				continue
			}
			glog.Infof("robot %s score no enough, kickout room", k)
			errcode := t.leave(k, 0)
			if errcode != pb.OK {
				continue
			}
			// 通知自己回到大厅
			ntf := &pb.ABLeaveNtf{
				Userid: k,
				Seat:   v.Seat,
			}
			t.send2userid(k, ntf)
			//离开状态消息
			t.userLeaveDesk(k)
		}
		for _, u := range t.NewbiewRobot {
			if u.Diamond >= int64(robot.LeaveLimit[0]) && u.Diamond <= int64(robot.LeaveLimit[1]) {
				// 分数区间正常
				continue
			}
			glog.Infof("newbiew robot %s score no enough, kickout room", u.Userid)
			// 归还头像
			t.returnHead(u)
			// 换个名字和分数
			u.Nickname = login.RandName()
			u.Photo = strconv.Itoa(utils.RandIntN(30) + 1)
			u.Diamond = int64(utils.RandMN(int(robot.InitScore[0]), int(robot.InitScore[1])))
			handler.RobotVip(u)
			// 自定义头像
			t.getHead(u)
		}
	}
}

// 踢除离线玩家
func (t *Desk) kickOffline() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		for k, v := range t.roles {
			if !v.Offline {
				continue
			}
			errcode := t.leave(k, 0)
			if errcode != pb.OK {
				continue
			}
			//离开状态消息
			t.userLeaveDesk(k)
		}
	}
}

// 剔除5局不下注的玩家
func (t *Desk) kickNoBet() {
	for k, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		if v.NoBetTimes < 5 {
			continue
		}
		errcode := t.leave(k, int32(pb.NoBet5Round))
		if errcode != pb.OK {
			continue
		}
		// 通知自己回到大厅
		ntf := &pb.ABLeaveNtf{
			Userid: k,
			Seat:   v.Seat,
		}
		t.send2userid(k, ntf)
		//离开状态消息
		t.userLeaveDesk(k)
		t.offlineMsg(k)
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
		// 通知自己回到大厅
		ntf := &pb.ABLeaveNtf{
			Userid: k,
			Seat:   v.Seat,
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

// 状态检测
func (t *Desk) stateCheck() {
	for k, role := range t.roles {
		if role.Robot {
			continue
		}
		kick := false
		var err pb.ErrCode
		switch t.Dtype {
		case int32(pb.DESK_TYPE_NEWBIEW):
			// 新手房间
			kick = role.State != data.NoveiceState && role.State != data.ExceptionState && role.State != data.FrothState
			err = pb.NewbieKick
		case int32(pb.DESK_TYPE_NORMAL):
			// 正常房间
			kick = role.State != data.NormalState || role.PCSwitch
			// kick = role.PCSwitch
		case int32(pb.DESK_TYPE_NORMAL_B):
			// 正常房间
			kick = role.State != data.NormalState || role.PCSwitch || role.RegistArea != 1
		}
		if kick {
			glog.Infof("user state change,dtype:%d,state:%d", t.Dtype, role.State)
			errcode := t.leave(k, int32(err))
			if errcode != pb.OK {
				continue
			}
			// 通知自己回到大厅
			ntf := &pb.ABLeaveNtf{
				Userid: k,
				Seat:   role.Seat,
			}
			t.send2userid(k, ntf)
			//离开状态消息
			t.userLeaveDesk(k)
			t.offlineMsg(k)
		}
	}
}

// pub房间人数为0时解散
func (t *Desk) checkPubOver() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		if !t.DeskData.Pub {
			//return
		}
	default:
		//return
	}
	if len(t.roles) != 0 {
		return
	}
	g := config.GetGame(t.DeskData.Unique)
	if g.Id == t.DeskData.Unique {
		return //配置房间不关闭
	}
	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

func (t *Desk) checkPubOver2() {
	switch t.state {
	case int32(pb.STATE_READY):
		t.closeTime++
		if t.closeTime == 60 {
			t.closeTime = 0
			t.checkPubOver()
		}
	default:
		t.closeTime = 0
	}
}

//.

// '结束游戏
// 是否过期
func (t *Desk) checkExpire() bool {
	var now = utils.Timestamp()
	if now > int64(t.DeskData.Expire) {
		glog.Debugf("game stop expire -> %d, %d",
			t.DeskData.Expire, now)
		return true
	}
	return false
}

// 是否结束游戏
func (t *Desk) checkOver() bool {
	if t.DeskData.Round == t.DeskGame.Round {
		glog.Debugf("game stop round -> %d, %d",
			t.DeskGame.Round, t.DeskData.Round)
		return true
	}
	return t.checkExpire()
}

// 结束牌局
func (t *Desk) gameStop() {
	if !t.checkOver() {
		return
	}

	//停止服务
	msg1 := new(pb.ServeStop)
	t.selfPid.Tell(msg1)
}

// 返还钻石
func (t *Desk) backCost() {
	//已经打过的房间不返还
	if t.DeskPriv == nil ||
		t.DeskGame.Round != 0 {
		return
	}
	//已经开始游戏不返还
	if t.state != int32(pb.STATE_FREE) &&
		t.state != int32(pb.STATE_READY) {
		glog.Errorf("game start priv room cost not back: %d", t.state)
		return
	}
	if t.DeskData.Cost <= 0 {
		return
	}
	//A
	if t.DeskData.Payment != 1 {
		// t.sendDiamond(t.DeskData.Cid,
		// int64(t.DeskData.Cost), int32(pb.LOG_TYPE3))
		t.sendCurrency(t.DeskData.Cid, 0, int64(t.DeskData.Cost), int32(pb.LOG_TYPE3), "AB娱乐对战房解散返还")
		return
	}
	//AA
	for k := range t.roles {
		// t.sendDiamond(k, int64(t.DeskData.Cost), int32(pb.LOG_TYPE3))
		t.sendCurrency(k, 0, int64(t.DeskData.Cost), int32(pb.LOG_TYPE3), "AB娱乐对战房解散返还")
	}
}

// 选择joker牌
func (t *Desk) choiceJoker() {
	t.shuffle()
	t.ABDeskFree.Joker = t.DeskGame.Cards[0] // joker
	t.DeskGame.Cards = t.DeskGame.Cards[1:]
	msg := &pb.ABFreeJokerNtf{
		Value: t.ABDeskFree.Joker,
	}
	t.broadcast4(msg)
}

// 选择赢家
/* func (t *Desk) winner(cardtype int32) (uint32, uint32) {
	platWinPro := 10000
	// cardTypes := t.Game.AB.CardType

	// for _, c := range cardTypes {
	// 	if cardtype == int32(c.Id) {
	// 		platWinPro = c.WinPro
	// 		break
	// 	}
	// }
	// maxWin := t.getMaxWinScore()

	wins := t.calWinningScoreOf16()
	// group := fmt.Sprintf("%d,%d", len(wins), 16-len(wins))

	// leadWeight := t.Game.AB.LeadWeights[len(t.Game.AB.LeadWeights)-1]
	// for _, l := range t.Game.AB.LeadWeights {
	// 	if l.Group == group {
	// 		leadWeight = l
	// 		break
	// 	}
	// }

	var lead DrawResult
	// 平台赢
	if utils.RandWan(int32(platWinPro)) && len(wins) > 0 || len(loses) <= 0 {
		glog.Infof("platform win, cardtype: %d,probabilty:%d,loses len:%d", cardtype, platWinPro, len(loses))
		weights := t.natureProbality(wins, leadWeight.PlatWinWeights, leadWeight.NatureWeights)
		index, err := utils.ChoiceIntIndex(weights)
		if err != nil {
			index = 0
		}
		lead = wins[index]
	} else {
		glog.Infof("platform lose,cardtype:%d, probabilty:%d", cardtype, platWinPro)
		weights := t.natureProbality(loses, leadWeight.PlatLoseWeights, leadWeight.NatureWeights)
		index, err := utils.ChoiceIntIndex(weights)
		if err != nil {
			index = len(loses) - 1
		}
		lead = loses[index]
	}

	return lead.Winner, lead.SideWinner
} */

func (t *Desk) check105(role *data.DeskRole) bool {
	// if role.Kick105Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 105*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) check200(role *data.DeskRole) bool {
	// if role.Kick200Flag {
	// 	return true
	// }
	// if role.OutDiamond >= 200*100 && role.Money == 0 {
	// 	return false
	// } else {
	// 	return true
	// }
	return true
}

func (t *Desk) checkGameTime(role *data.DeskRole) bool {
	if role.State == 3 {
		return true
	}
	bean := table.GetTables().NewbieTable.Get()
	if bean == nil {
		return true
	}

	if role.TotalGameTime >= uint64(bean.Transfer[0])*60 && role.Money == 0 {
		return false
	} else {
		return true
	}
}

func (t *Desk) checkCloseServer() bool {
	if t.closeServer {
		return false
	} else {
		return true
	}

}

func (t *Desk) check5RoundNoBet(role *data.DeskRole) bool {
	if role.GetRobot() {
		return true
	}
	if role.NoBetTimes >= 100 {
		return false
	}
	return true
}

// 私人房进入房间响应消息
func (t *Desk) privEnterMsg() *pb.ABEnterRoomRsp {
	msg := new(pb.ABEnterRoomRsp)
	msg.Gameid = t.Game.Id
	msg.Gmode = t.Gmode
	msg.Roomid = t.Rid
	msg.GameType = t.Gtype
	//房间数据
	msg.Roominfo = handler.PackABPrivRoom(t.DeskData)
	msg.Roominfo.State = t.state
	msg.Roominfo.Dealer = t.Dealer
	msg.Roominfo.DealerSeat = t.DealerSeat
	msg.Roominfo.CurRound = t.DeskGame.Round
	if t.ABDeskPriv != nil {
		msg.LeftSeat = t.privLeftSeats() // 剩余座位号

		msg.Roominfo.Poker = &pb.ABPrivRoomOver{
			ValueA: t.ABDeskPriv.AndarCards,
			ValueB: t.ABDeskPriv.BaharCards,
			Joker:  t.ABDeskPriv.Joker,
		}
		if t.state == int32(pb.STATE_BET) {
			now := utils.Timestamp()
			msg.Roominfo.Timer = uint32(t.Game.AB.BetTime) - uint32(now-t.CardRoundTime)
		}

		if t.state == int32(pb.STATE_CHARGE) {
			msg.Roominfo.Timer = uint32(ChargeInGameTime - t.timer)
			msg.Roominfo.Totaltimer = ChargeInGameTime
			msg.RechargeTime = int64(ChargeInGameTime-t.timer) + utils.LocalTime().Unix()
			msg.ChargingSeats = t.chargingSeats
			if len(t.chargingSeats) > 0 {
				amount, give := t.getRechargeAmount(t.chargingSeats[0])
				msg.Amount, msg.GiveAmount = int32(amount), int32(give)
			}
		}

		// 下注总额
		andar, bahar := t.getPrivAndarBaharBets()
		msg.Roominfo.Bets = append(msg.Roominfo.Bets, &pb.ABPrivUserBet{Seat: 1, Bets: andar})
		msg.Roominfo.Bets = append(msg.Roominfo.Bets, &pb.ABPrivUserBet{Seat: 2, Bets: bahar})
	}
	//坐下玩家信息
	msg.Userinfo = t.privSeatBetsMsg()
	//投票信息
	msg.Voteinfo = t.voteInfoMsg()
	return msg
}

// 所有坐下玩家数据
func (t *Desk) privSeatBetsMsg() (msg []*pb.ABPrivUser) {
	for _, v := range t.roles {
		msg2 := t.privSeatRoleMsg(v.Userid)
		if msg2 == nil {
			continue
		}
		msg = append(msg, msg2)
	}
	return
}

// 位置上玩家数据
func (t *Desk) privSeatRoleMsg(userid string) (msg *pb.ABPrivUser) {
	var user *data.User
	role, ok := t.roles[userid]
	if !ok || role == nil {
		return
	}
	user = role.User

	msg = handler.PackABPrivUser(user)
	msg.Seat = role.Seat
	msg.Fraction = user.GetFraction()
	msg.Ready = true
	msg.Offline = role.Offline
	// 下注信息
	msg.Bets = t.privUserSeatBetMsg(userid)
	return
}

// 玩家下注总额
func (t *Desk) privUserSeatBetMsg(userid string) []*pb.ABPrivUserBet {
	andar := &pb.ABPrivUserBet{Seat: 1, Bets: 0}
	bahar := &pb.ABPrivUserBet{Seat: 2, Bets: 0}
	bets := []*pb.ABPrivUserBet{andar, bahar}
	if t.ABDeskPriv == nil {
		return bets
	}
	andar.Bets = t.DeskPriv.AndarBets[userid]
	bahar.Bets = t.DeskPriv.BaharBets[userid]
	return bets
}

// 进入消息
func (t *Desk) privCameinMsg(userid string) {
	msg := new(pb.ABPrivCameinNtf)
	msg.Userinfo = t.privSeatRoleMsg(userid)
	t.broadcast4(msg)
}

const (
	PAUSE_REASON_0 = iota
	PAUSE_REASON_1 //发第一轮牌动画
	PAUSE_REASON_2 //发第二轮牌动画
	PAUSE_REASON_3 //
	PAUSE_REASON_4 //结算动画
)

// 暂停游戏
func (t *Desk) pauseGame(reason int, time int, arg interface{}) {
	t.state = int32(pb.STATE_PAUSE)
	t.timer = 0
	t.pauseTime = time
	t.reason = reason
	t.pauseArg = arg
}

// 恢复游戏
func (t *Desk) resumeGame(reason int) {
	switch reason {
	case PAUSE_REASON_1: // 第一轮牌发完
		// 通知开始下注
		t.privStartBet()

	case PAUSE_REASON_2: // 第二轮牌发完
		// 结算
		glog.Infof("ab priv room game over %d win", t.ABDeskPriv.Winner)
		t.privGameOver()

	case PAUSE_REASON_4:
		// 结算动画完毕
		t.state = int32(pb.STATE_READY) //设置房间状态
		t.pushState()

		switch t.Rtype {
		case int32(pb.ROOM_TYPE1):
			// 回合结束踢出离线超时玩家
			t.kickPrivOfflineTimeout()

			// 回合结束解散牌桌
			if t.dismissOnRoundOver {
				t.pushPrivSettle()
				return
			}

			if t.DeskGame.Round < uint32(t.DeskData.Round) {
				if t.Gmode == 0 {
					var chargings []uint32
					for userId, role := range t.roles {
						if role.User.GetScore() < int64(t.Game.Min_Access) {
							seatId := t.getSeat(userId)
							chargings = append(chargings, seatId)
						}
					}
					// 需要充值的玩家
					if len(chargings) > 0 {
						t.chargeInGameBegin(chargings...)
						return
					}
				}
				// 对战自动开始下一轮
				if err := t.privGameStart(); err != pb.OK {
					// 下一轮失败,结算
					glog.Infof("start failed, game round over: %v, %v, %v", t.DeskGame.Round, t.DeskData.Round, err)
					t.pushPrivSettle()
				}
			} else {
				t.pushPrivSettle()
				glog.Infof("game round over: %v, %v", t.DeskGame.Round, t.DeskData.Round)
			}
		}

	default:
		// 真实玩家换桌
		glog.Error("unknown pause reason %v", reason)
	}
}

// 推送私人房回合结束结算消息
func (t *Desk) pushPrivSettle() {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return
	}

	// 展示结算页30s后退出房间
	t.settleExitTime = utils.Timestamp() + PrivSettleExitDelay
	// 推送结算页
	ntf := &pb.PrivSettleNtf{
		Code:           t.Code,
		Gtype:          t.Gtype,
		Time:           time.Unix(int64(t.Ctime), 0).Format(utils.FORMAT2),
		ExitTime:       PrivSettleExitDelay,
		RoundMinAccess: int64(t.DeskData.Game.Min_Access),
	}

	for userid, player := range t.DeskPriv.PrivPlayer {
		var again bool
		switch t.Gmode {
		case 0: // 真金
			if role, ok := t.roles[userid]; ok {
				again = role.GetScore() >= int64(t.DeskData.Game.Min_Access)
			}
		case 1: // 娱乐
			again = true
		}
		vipLv, _ := strconv.Atoi(player[2])
		ntf.Settles = append(ntf.Settles, &pb.PrivSettle{
			Userid:   userid,
			Nickname: player[0],
			Photo:    player[1],
			VipLv:    int32(vipLv),
			Fraction: t.PrivScore[userid],
			Wins:     t.PrivWins[userid],
			Loses:    t.PrivLoses[userid],
			Again:    again,
		})
	}
	t.broadcast(ntf)
}

// 检查可提现彩金
func (t *Desk) checkOutDiamond(userid string, score_final int64) {
	role := t.getRole(userid)
	if role == nil {
		return
	}
	if role.Robot {
		return
	}

	var incre int64
	if score_final > 0 { //赢
		old := role.OutDiamond               //旧可提彩金
		new := role.OutDiamond + score_final //新可提彩金
		if new > role.Diamond {              //可提彩金大于携带彩金
			new = role.Diamond //可提彩金不能超过携带彩金
		}
		incre = new - old //可提彩金增量
	} else if score_final < 0 { //输
		old := role.OutDiamond
		new := role.OutDiamond
		if role.OutDiamond > role.Diamond { //可提彩金不能超过携带彩金
			new = role.Diamond
		}
		incre = new - old //可提彩金增量
	}

	role.OutDiamond += incre //修改本地可提彩金数据

	//同步变化
	msg := &pb.GiveAndOutCash{
		Userid: userid,
		GameId: t.GameId,
		Desc:   fmt.Sprintf("房间%s", t.Rid),
	}
	msg.Out = incre
	if role.Offline {
		t.rolePid.Tell(msg)
	} else {
		t.send2userid(userid, msg)
	}
}

func (t *Desk) natureProbality(draws []DrawResult, weights []int, nature []int) []int {
	if len(draws) > len(weights) {
		glog.Errorf("config is err,draw result length > weight length")
		return weights
	}
	reslut := make([]int, 0)
	for i, d := range draws {
		w := weights[i]
		var addition int = 0
		if d.Winner == 1 {
			// andar
			addition = nature[(d.SideWinner-3)*2]
		} else {
			// bahar
			addition = nature[(d.SideWinner-2)*2-1]
		}
		reslut = append(reslut, w+addition) // 最终的权重
	}
	return reslut
}

func (t *Desk) getPokerIntervalBySeat(seat uint32) []int {
	Indexs := make([]int, 0)
	switch seat {
	case 1:
		// andar
		for i := 1; i <= 48; i++ {
			if i%2 != 0 {
				Indexs = append(Indexs, i)
			}
		}
	case 2:
		// bahar
		for i := 1; i <= 48; i++ {
			if i%2 == 0 {
				Indexs = append(Indexs, i)
			}
		}
	case 3:
		for i := 1; i <= 5; i++ {
			Indexs = append(Indexs, i)
		}
	case 4:
		for i := 6; i <= 10; i++ {
			Indexs = append(Indexs, i)
		}
	case 5:
		for i := 11; i <= 15; i++ {
			Indexs = append(Indexs, i)
		}
	case 6:
		for i := 16; i <= 25; i++ {
			Indexs = append(Indexs, i)
		}
	case 7:
		for i := 26; i <= 30; i++ {
			Indexs = append(Indexs, i)
		}
	case 8:
		for i := 31; i <= 35; i++ {
			Indexs = append(Indexs, i)
		}
	case 9:
		for i := 36; i <= 40; i++ {
			Indexs = append(Indexs, i)
		}
	default:
		for i := 41; i <= 51; i++ {
			Indexs = append(Indexs, i)
		}
	}

	return Indexs
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
