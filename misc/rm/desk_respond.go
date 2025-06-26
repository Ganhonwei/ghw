package main

import (
	"fmt"
	"strings"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// external function

// ' 打印
func (t *Desk) printOver() {
	glog.Debugf("game over data -> %#v", t.DeskData)
	glog.Debugf("game over roles -> %d", len(t.roles))
	glog.Debugf("game over router -> %d", len(t.router))
	glog.Debugf("game over state -> %d", t.state)
	glog.Debugf("game over dealer -> %s", t.DeskGame.Dealer)
	glog.Debugf("game over dealer seat -> %d", t.DeskGame.DealerSeat)
	for k, v := range t.roles {
		glog.Debugf("game over userid %s -> %d", k, v.Seat)
	}
	for k, v := range t.seats {
		glog.Debugf("game over seat %d -> %s", k, v.Userid)
	}
	if t.isFree() {
		glog.Debugf("game over Carry -> %d", t.DeskFree.Carry)
		glog.Debugf("game over DealerNum -> %d", t.DeskFree.DealerNum)
		glog.Debugf("game over Cards -> %#v", t.DeskFree.Cards)
		glog.Debugf("game over Bets -> %#v", t.DeskFree.Bets)
		glog.Debugf("game over SeatBets -> %#v", t.DeskFree.SeatBets)
		glog.Debugf("game over Multiple -> %#v", t.DeskFree.Multiple)
		glog.Debugf("game over Score1 -> %#v", t.DeskFree.Score1)
		glog.Debugf("game over Score2 -> %#v", t.DeskFree.Score2)
		glog.Debugf("game over Score3 -> %#v", t.DeskFree.Score3)
	}
}

// 聊天
func (t *Desk) chatText(arg *pb.ChatTextReq, ctx actor.Context) {
	userid := t.getRouter(ctx)
	seat := t.getSeatid(userid)
	glog.Debugf("ChatTextReq %s, %d", userid, seat)
	//房间消息广播,聊天
	msg := handler.ChatTextMsg(seat, userid, arg.Content)
	// if role, ok := t.roles[userid]; ok {
	// 	// 扣钱
	// 	vip := table.GetTables().VipTable.Get(int32(role.Vip.Lv))
	// 	if vip.EmojiCost < 0 {
	// 		msg.Error = pb.VipLvTooLow
	// 		t.send2userid(userid, msg)
	// 		return
	// 	}
	// 	if role.Diamond-int64(vip.EmojiCost) < int64(t.DeskData.Game.RM.Bottom)*80 {
	// 		msg.Error = pb.NotEnoughCoin
	// 		t.send2userid(userid, msg)
	// 		return
	// 	}
	// 	if vip.EmojiCost > 0 {
	// 		t.sendCurrency(userid, -int64(vip.EmojiCost), int32(pb.LOG_TYPE102), "rm发表情")
	// 	}
	// }
	// msg.Error = t.chatEmoji(userid, arg.GetContent())
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	t.send2userid(userid, msg)

	msg1 := &pb.ChatTextNtf{From: seat, To: arg.To, Content: arg.Content}
	t.broadcast(msg1)
}

// func (t *Desk) chatVoice(arg *pb.ChatVoiceReq, ctx actor.Context) {
// 	userid := t.getRouter(ctx)
// 	seat := t.getSeatid(userid)
// 	glog.Debugf("ChatVoiceReq %s, %d", userid, seat)
// 	//房间消息广播,聊天
// 	t.broadcast(handler.ChatVoiceMsg(seat, userid, arg.Content))
// }

// 表情包,TODO 严格验证或新加协议
func (t *Desk) chatEmoji(userid, context string) pb.ErrCode {
	if !strings.HasPrefix(context, "_p") {
		return pb.OK
	}
	s := strings.Split(context, "/")
	if len(s) != 3 {
		return pb.OK
	}
	var num int64 = 5 //TODO 消耗数量配置
	if s[1] == "-1" {
		n := len(t.seats) - 1
		if n > 0 {
			num *= int64(n)
		}
	}
	if v, ok := t.roles[userid]; ok && v != nil {
		if v.User.GetDiamond() < num {
			return pb.NotEnoughDiamond
		}
	}
	//TODO 货币变更消息广播
	t.sendDiamond(userid, -1*num, int32(pb.LOG_TYPE50))
	return pb.OK
}

// enter 进入桌子
func (t *Desk) enter(user *data.User, pid *actor.PID) pb.ErrCode {
	errcode := t.enterCheck(user)
	if errcode != pb.OK {
		return errcode
	}
	//设置路由
	t.router[pid.String()] = user.GetUserid()
	// 已经在房间防止数据覆盖
	if v, ok := t.roles[user.GetUserid()]; ok {
		v.Offline = false
		v.Pid = pid
		//在线消息
		t.offlineMsg(user.GetUserid())
		return pb.AlreadyInRoom //已经在房间内
	}
	//加入游戏
	p := new(data.DeskRole)
	p.User = user
	p.Pid = pid
	//私人,自由场分配位置
	t.enterSeat(p)
	t.roles[user.GetUserid()] = p
	return pb.OK
}

// 进入桌子分配位置
func (t *Desk) enterSeat(p *data.DeskRole) {
	if p.Seat != 0 {
		return
	}
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		return
	}
	idleSeats := make([]int, 0)
	for k := range t.FakeSeats {
		if _, ok := t.seats[k]; !ok {
			idleSeats = append(idleSeats, int(k))
		}
	}
	seats := make(map[uint32]string)
	for k := range t.seats {
		seats[k] = ""
	}
	i := handler.GankPartner(p.Robot, t.DeskData.Count, seats)
	// var i uint32
	// if len(idleSeats) > 0 {
	// 	winner, err := utils.ChoiceInt(idleSeats)
	// 	if err == nil {
	// 		i = uint32(winner)
	// 	}
	// } else {
	// 	i = uint32(utils.RandInt32N(int32(t.DeskData.Count)) + 1) // 先随机一个位置
	// 	var j uint32
	// 	for j = 1; j <= t.DeskData.Count; j++ {
	// 		if _, ok := t.seats[i]; !ok {
	// 			break
	// 		} else {
	// 			i++
	// 			if i >= t.DeskData.Count {
	// 				i = 1
	// 			}
	// 		}
	// 	}
	// }
	p.Seat = i
	t.seats[i] = &data.DeskSeat{
		Userid: p.User.GetUserid(),
	}
	if t.IsGaming() {
		t.seats[i].Watch = true
	}
	// var i uint32
	// for i = 1; i <= t.DeskData.Count; i++ {
	// 	if _, ok := t.seats[i]; !ok {
	// 		p.Seat = i
	// 		t.seats[i] = &data.DeskSeat{
	// 			Userid: p.User.GetUserid(),
	// 		}
	// 		if t.IsGaming() {
	// 			t.seats[i].Watch = true
	// 		}
	// 		break
	// 	}
	// }
}

// TODO 不同类型房间进入限制
func (t *Desk) enterCheck(user *data.User) pb.ErrCode {
	//if user.GetVip() < t.DeskData.Vip {
	//	return pb.VipTooLow
	//}
	//if user.GetChip() < int64(t.DeskData.Chip) {
	//	return pb.ChipNotEnough
	//}
	if _, ok := t.roles[user.GetUserid()]; ok {
		return pb.OK // 已在房间中
	}

	if uint32(len(t.roles)) >= t.DeskData.Count {
		return pb.RoomFull //人数已满
	}
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		if t.IsGaming() || t.DeskGame.Round > 0 || t.AgainRound > 0 {
			return pb.GameStarted // 已经开始不能加入
		}
		//可以旁观
		// default:
		// 	if uint32(len(t.roles)) >= t.DeskData.Count {
		// 		return pb.RoomFull //人数已满
		// 	}
	}
	//消耗
	if t.isAADesk() && user.GetUserid() != t.DeskData.Cid {
		if user.GetDiamond() < int64(t.DeskData.Cost) {
			return pb.NotEnoughDiamond
		}
	}
	// if user.GetCoin() < t.DeskData.Maximum {
	// 	return pb.NotEnoughCoin
	// }
	return pb.OK
}

//.

//' 离开房间处理

// nnLeave 玩家主动离开,TODO 下注也可以离开
func (t *Desk) nnLeave(userid string, ctx actor.Context) {
	//离线
	defer t.offlineDesk(userid)
	//seat := t.getSeat(userid)
	errcode := t.leave(userid)
	t.notifyNodeLeaveEarly(t.Rid, userid, errcode)
	t.notifyGateUserLeft(userid, errcode, 0)
	if errcode == pb.OK || errcode == pb.LeaveEarly {
		//清除数据
		t.userLeaveDesk(userid)
		return
	}
	msg := new(pb.RMLeaveRsp)
	//TODO 暂时不返回错误
	msg.Error = errcode
	ctx.Respond(msg)
}

// 通知节点玩家提前离开
func (t *Desk) notifyNodeLeaveEarly(roomdid, userid string, err pb.ErrCode) {
	if err == pb.LeaveEarly {
		msg := &pb.EarlyLeave{
			Roomid: roomdid,
			Userid: userid,
		}
		nodePid.Tell(msg)
	}
}

// 通知网关玩家离开
func (t *Desk) notifyGateUserLeft(userid string, err pb.ErrCode, reason int32) {
	if v, ok := t.roles[userid]; ok && v != nil {
		if !v.Offline ||
			t.Rtype == int32(pb.ROOM_TYPE1) { // 保证私人房解散即将离开时掉线也能同步到 gate
			msg := new(pb.LeftDesk)
			msg.Error = err
			msg.Gtype = int32(pb.RUMMY)
			msg.Reason = reason
			v.Pid.Tell(msg)
		}
	}
}

// leave 玩家离开,TODO 下注也可以离开
func (t *Desk) leave(userid string) pb.ErrCode {
	//离线处理
	// defer t.userLeaveDeskMsg(userid)
	// if _, ok := t.roles[userid]; !ok {
	// 	return pb.NotInRoom
	// }
	seatid := t.getSeatid(userid)
	role := t.getRole(userid)
	seat := t.getSeat(seatid)
	if role == nil || seat == nil {
		return pb.NotInRoom
	}

	//离开房间条件
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		if t.IsGaming() {
			if seat.Ready {
				status := t.getStatus(seatid)
				if status == nil {
					return pb.NotInRoom
				}
				if status.Alive {
					return pb.GameStartedCannotLeave
				} else {
					return pb.LeaveEarly
				}
			} else if seat.Watch {
				return pb.LeaveEarly
			} else {
				return pb.OK
			}
		} else {
			return pb.OK
		}
		// switch t.state {
		// case int32(pb.STATE_READY):
		// default:
		// 	if deskSeat.Ready {
		// 		if status.Alive {
		// 			return pb.GameStartedCannotLeave
		// 		} else {
		// 			role.Leave = true
		// 			return pb.LeaveEarly
		// 		}
		// 	}
		// }
		// case int32(pb.ROOM_TYPE1): //私人
		// 	if t.DeskPriv.VoteSeat != 0 {
		// 		return pb.VotingCantLaunchVote
		// 	}
		// 	switch t.state {
		// 	case int32(pb.STATE_READY):
		// 	default:
		// 		seat := t.getSeat(userid)
		// 		if v, ok := t.seats[seat]; ok {
		// 			if v.Ready {
		// 				return pb.GameStartedCannotLeave
		// 			}
		// 		}
		// 	}
		// case int32(pb.ROOM_TYPE2): //百人
		// 	//清空上庄列表
		// 	t.leaveBeDealer(userid)
		// 	switch t.state {
		// 	case int32(pb.STATE_BET): //下注中
		// 		if _, ok := t.DeskFree.Bets[userid]; ok {
		// 			return pb.GameStartedCannotLeave
		// 		}
		// 		if userid == t.DeskGame.Dealer {
		// 			return pb.GameStartedCannotLeave
		// 		}
		// 	default:
		// 		//庄家下庄
		// 		user := t.getPlayer(userid)
		// 		t.delBeDealer(userid, user)
		// 	}
	}
	return pb.OK
}

// 离开状态消息
func (t *Desk) userLeaveDeskMsg(userid string) {
	if v, ok := t.roles[userid]; ok {
		if !v.Offline ||
			t.Rtype == int32(pb.ROOM_TYPE1) { // 保证私人房解散即将离开时掉线也能同步到 gate
			msg2 := new(pb.LeftDesk)
			v.Pid.Tell(msg2)
		}
		//
		//msg := new(pb.LeaveDesk)
		//msg.Userid = userid
		//msg.Roomid = t.DeskData.Rid
		//离开房间
		//nodePid.Tell(msg)
		//离开房间,TODO 玩家进程中操作,一致性
		//t.roomPid.Request(msg, t.selfPid)
	}
}

// 离开状态消息
func (t *Desk) userLeaveDesk(userid string, args ...pb.ErrCode) {
	err := pb.OK
	if len(args) != 0 {
		err = args[0]
	}
	// 私人房提前离开回到大厅不返回ok
	if t.Rtype == int32(pb.ROOM_TYPE1) && err == pb.OK {
		err = pb.LeaveEarly
	}
	if v, ok := t.roles[userid]; ok {
		//离开状态消息
		msg1 := new(pb.RMLeaveRsp)
		msg1.Seat = v.Seat
		msg1.Userid = userid
		t.send2userid(userid, msg1)

		//广播
		msg2 := new(pb.RMLeaveNtf)
		msg2.Seat = v.Seat
		msg2.Userid = userid
		msg2.Error = err

		t.broadcast(msg2)

		//msg2 := new(pb.LeftDesk)
		//v.Pid.Tell(msg2)
		//清除数据

		delete(t.seats, v.Seat)
		delete(t.roles, userid)

		//房间状态检查
		if t.state == int32(pb.STATE_READY) && t.roleNum() < 2 {
			t.state = int32(pb.STATE_FREE)
			t.pushState()
		}

		//
		msg := new(pb.LeaveDesk)
		msg.Userid = userid
		msg.Roomid = t.DeskData.Rid
		//离开房间
		nodePid.Tell(msg)

		//离开房间,TODO 玩家进程中操作,一致性
		t.roomPid.Request(msg, t.selfPid)
	}
}

func (t *Desk) offlineMsg(userid string) {
	if v, ok := t.roles[userid]; ok && v != nil {
		msg := new(pb.RMPushOfflineNtf)
		msg.Seat = v.Seat
		msg.Userid = userid
		msg.Offline = v.Offline
		t.broadcast2(userid, msg)
	}
}

//.

//' 庄家操作

//暂时同一个玩家上庄列表只能有一个
//上庄规则是最高携带优先

// 上庄或补庄处理，st:0下庄 1上庄 2补庄
func (t *Desk) addBeDealer(userid string, st int32,
	num int64, user *data.User) {
	switch st {
	case int32(pb.DEALER_BU): //庄家补庄
	case int32(pb.DEALER_UP): //庄家上庄
	default:
		return
	}
	if userid == t.DeskGame.Dealer {
		t.DeskFree.Carry += num
	} else {
		t.DeskFree.Dealers[userid] += num
	}
	msg := handler.BeRMDealerMsg(st, num, t.DeskGame.Dealer,
		userid, user.GetNickname(), user.GetPhoto())
	t.broadcast(msg)
}

// 下庄处理
func (t *Desk) delBeDealer(userid string, user *data.User) {
	if !t.isFreeDealer() {
		return
	}
	if t.DeskGame.Dealer != userid {
		return
	}
	var num int64
	num = t.DeskFree.Carry
	t.sendCoin(userid, num, int32(pb.LOG_TYPE8))
	t.DeskGame.Dealer = ""
	t.DeskGame.DealerSeat = 0
	t.DeskFree.Carry = 0
	t.DeskFree.DealerNum = 0
	msg := handler.BeRMDealerMsg(0, num, t.DeskGame.Dealer,
		userid, user.GetNickname(), user.GetPhoto())
	t.broadcast(msg)
}

// 离开房间清空上庄列表
func (t *Desk) leaveBeDealer(userid string) {
	if !t.isFree() {
		return
	}
	delete(t.DeskFree.Dealers, userid)
}

// 开始游戏时选择玩家成为庄家
func (t *Desk) beComeDealer() {
	//if !t.isFreeDealer() {
	//	return
	//}
	userid, num := t.findBeDealer()
	if userid == "" {
		//glog.Errorf("beComeDealer failed %s, %d", userid, num)
		return
	}
	//玩家站起
	t.roleSitUp(userid)
	//上庄成功扣除
	t.sendCoin(userid, (-1 * num), int32(pb.LOG_TYPE7))
	//成为庄家
	t.DeskGame.Dealer = userid
	t.DeskFree.Carry = num
	t.DeskFree.DealerNum = 0
	delete(t.DeskFree.Dealers, userid)
	//消息
	user := t.getPlayer(userid)
	if user == nil {
		return
	}
	msg := handler.BeRMDealerMsg(1, num, t.DeskGame.Dealer,
		userid, user.GetNickname(), user.GetPhoto())
	t.broadcast(msg)
}

// 携带最大的优先做庄
func (t *Desk) findBeDealer() (userid string, carry int64) {
	if !t.isFree() {
		return
	}
	for k, v := range t.DeskFree.Dealers {
		if !t.isOnline(k) || v < int64(t.DeskData.Carry) {
			delete(t.DeskFree.Dealers, k)
			continue
		}
		if val, ok := t.roles[k]; ok {
			//自动下庄金额不足玩家
			if val.GetCoin() < int64(t.DeskData.Carry) {
				delete(t.DeskFree.Dealers, k)
				continue
			}
			//全部资金上庄
			//if val.GetCoin() > carry {
			//	userid = k
			//	carry = val.GetCoin()
			//}
			//选择金额上庄
			if v > carry && val.GetCoin() >= v {
				userid = k
				carry = v
			}
		}
	}
	return
}

// 是否百人上庄
func (t *Desk) isFree() bool {
	if !t.DeskData.Deal {
		return false
	}
	if t.DeskFree == nil {
		return false
	}
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
	default:
		return false
	}
	return true
}

// 是否百人有庄
func (t *Desk) isFreeDealer() bool {
	if !t.isFree() {
		return false
	}
	if t.DeskGame.Dealer == "" {
		return false
	}
	return true
}

// 结束时检测不足做庄
func (t *Desk) checkBeDealer() {
	if !t.isFree() {
		return
	}
	//做庄次数
	t.DeskFree.DealerNum++
	if v, ok := t.roles[t.DeskGame.Dealer]; ok {
		//离线或者不足
		if v.Offline || t.leftDealerTimes() == 0 ||
			t.DeskFree.Carry <= int64(t.DeskData.Down) ||
			t.DeskFree.Carry >= int64(t.DeskData.Top) ||
			t.DeskFree.DealerDown {
			t.DeskFree.DealerDown = false
			t.delBeDealer(t.DeskGame.Dealer, v.User)
		} else {
			return
		}
	}
	//选择成为庄家
	t.beComeDealer()
	//无人坐庄
	if t.DeskGame.Dealer == "" {
		//庄家每次都补庄
		if t.DeskFree.Carry < SysCarry {
			t.DeskFree.Carry = SysCarry
		}
		//重置次数
		if t.leftDealerTimes() == 0 {
			t.DeskFree.DealerNum = 0
		}
	}
}

// 是否已经是庄家或者已经申请上庄
func (t *Desk) alreadyBeDealer(userid string) bool {
	if !t.isFree() {
		return false
	}
	//已经是庄
	if t.DeskGame.Dealer == userid {
		return true
	}
	//已经申请
	if _, ok := t.DeskFree.Dealers[userid]; ok {
		return true
	}
	return false
}

func (t *Desk) readying2(userid string) {
	msg := new(pb.RMReady2Rsp)

	if t.state == int32(pb.STATE_DEALING) {
		seat := t.getSeatid(userid)
		if v, ok := t.seats[seat]; ok {
			v.Ready2 = true
		}
		msg.Error = pb.OK
	} else {
		msg.Error = pb.OperateError
	}

	t.send2userid(userid, msg)

	if t.state == int32(pb.STATE_DEALING) && t.allReady2() {
		t.state = int32(pb.STATE_BET) //切换状态为下注状态
		t.pushState()
		t.initAct()
	}
}

// . 摸牌
func (t *Desk) drawCard(userid string, area uint32, timeout bool) {
	msg := new(pb.RMDrawCardRsp)
	msg1 := new(pb.RMDrawCardNtf)

	if t.state != int32(pb.STATE_BET) {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	seatid := t.getSeatid(userid)

	if t.ActSeat != seatid {
		msg.Error = pb.NotYourTurn
		t.send2userid(userid, msg)
		return
	}

	seat := t.getSeat(seatid)
	role := t.getRole(userid)
	if seat == nil || role == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if len(seat.Cards) != 13 {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if area == 1 { //摸牌堆
		if len(t.DeskGame.Cards) == 0 { //摸牌堆没牌了
			msg.Error = pb.Failed
			t.send2userid(userid, msg)
			return
		}

		id := t.cardTypeId
		ct := t.Game.RM.FindCardType(id)
		robot := role.Robot

		var choices []utils.Choice
		if !robot {
			for k, v := range ct.PlayerDrawCardWeight {
				choices = append(choices, utils.Choice{Weight: v, Item: k}) //0随机、1边张、2joker
			}
		} else {
			for k, v := range ct.RobotDrawCardWeight {
				choices = append(choices, utils.Choice{Weight: v, Item: k})
			}
		}
		c, _ := utils.WeightedChoice(choices)

		op := c.Item.(int)

		if !robot && t.isTriggerMustLoseControl() { //必输局玩家处理
			op = 3 //发非边张非joker牌
			if !t.isTriggerMustLoseDrawCard {
				t.isTriggerMustLoseDrawCard = true
			}
		}

		if robot && seat.Core && t.isTriggerMustLoseControl() { //必输局核心人机处理
			if utils.RandWan(int32(t.mustLoseCoreRobotGoodCardProb)) {
				if op != 1 && op != 2 {
					choices = []utils.Choice{}
					choices = append(choices, utils.Choice{Weight: 70, Item: 1}) //边张
					if algo.GetRMLaiziNum(seat.Cards, t.WildCard) < 6 {
						choices = append(choices, utils.Choice{Weight: 30, Item: 2}) //joker
					}
					c, _ = utils.WeightedChoice(choices)
					op = c.Item.(int)
					if !t.isTriggerMustLoseCoreRobotGoodCard {
						t.isTriggerMustLoseCoreRobotGoodCard = true
					}
				}
			}
		}

		if op == 1 { //边张牌
			card, remain_cards := algo.DealRandomBianZhangCard(t.DeskGame.Cards, seat.Cards)
			t.DeskGame.Cards = remain_cards
			seat.Cards = append(seat.Cards, card)
			msg.Card = card
			// msg1.Nextqicard = card
		} else if op == 2 { //joker牌
			if algo.GetRMLaiziNum(seat.Cards, t.WildCard) < 6 { //最多6张癞子
				card, remain_cards := algo.DealRandomJoker(t.DeskGame.Cards, t.WildCard)
				t.DeskGame.Cards = remain_cards
				seat.Cards = append(seat.Cards, card)
				msg.Card = card
			} else { //超过6张发非joker随机牌
				card, remain_cards := algo.DealRandomExceptJoker(t.DeskGame.Cards, t.WildCard)
				t.DeskGame.Cards = remain_cards
				seat.Cards = append(seat.Cards, card)
				msg.Card = card
			}
			// msg1.Nextqicard = card
		} else if op == 3 { //非边张非joker随机牌
			card, remain_cards := algo.DealRandomExceptBianZhangCardAndJoker(t.DeskGame.Cards, seat.Cards, t.WildCard)
			t.DeskGame.Cards = remain_cards
			seat.Cards = append(seat.Cards, card)
			msg.Card = card
		} else { //随机牌
			card, remain_cards := algo.DealRandomCard(t.DeskGame.Cards)
			t.DeskGame.Cards = remain_cards
			seat.Cards = append(seat.Cards, card)
			msg.Card = card
			// msg1.Nextqicard = card
		}

	} else if area == 2 { //摸弃牌堆
		if len(t.DeskGame.QiCards) == 0 { //弃牌堆没牌了
			msg.Error = pb.Failed
			t.send2userid(userid, msg)
			return
		}

		len1 := len(t.DeskGame.QiCards)
		card := t.DeskGame.QiCards[len1-1]
		t.DeskGame.QiCards = t.DeskGame.QiCards[:len1-1]
		seat.Cards = append(seat.Cards, card)

		msg.Card = card
		msg1.Qicard = card

		len2 := len(t.DeskGame.QiCards)
		if len2 == 0 {
			msg1.Nextqicard = 0
		} else {
			msg1.Nextqicard = t.DeskGame.QiCards[len2-1]
		}
	} else {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if !timeout {
		t.send2userid(userid, msg)
	} else {
		msg1.Card = msg.Card
	}

	msg1.Seat = seatid
	msg1.Area = area
	msg1.Left = uint32(len(t.DeskGame.Cards))

	if IsDebug && role.Robot { //调试模式广播人机摸得牌
		msg1.Card = msg.Card
	}

	t.broadcast(msg1)

	rm_detail := t.detail.FindRMDetail(seatid)
	if rm_detail != nil {
		rm_detail.MoCards = append(rm_detail.MoCards, msg.Card)
	}
}

// 出牌
func (t *Desk) discard(userid string, card uint32, timeout bool) {
	msg := new(pb.RMDiscardRsp)
	msg1 := new(pb.RMDiscardNtf)

	seatid := t.getSeatid(userid)

	if t.ActSeat != seatid {
		msg.Error = pb.NotYourTurn
		t.send2userid(userid, msg)
		return
	}

	seat := t.getSeat(seatid)
	if seat == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if len(seat.Cards) != 14 {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var new []uint32
	found := false
	for _, v := range seat.Cards {
		if !found && v == card {
			found = true
			continue
		}
		new = append(new, v)
	}

	if found {
		seat.Cards = new
		msg.Card = card
	} else {
		msg.Error = pb.Failed
		msg.Cards = seat.Cards
		t.send2userid(userid, msg)
		return
	}

	card = t.mustLoseRobotDiscardLimit(card) //必输局上家出牌限制

	t.DeskGame.QiCards = append(t.DeskGame.QiCards, card)

	if !timeout {
		t.send2userid(userid, msg)
	}

	msg1.Seat = seatid
	msg1.Card = card
	t.broadcast(msg1)

	seat.Sort = false //重置sort标记

	if !timeout {
		t.setNextActSeat()
	} else {
		t.pauseGame(PAUSE_REASON_1, 3, nil)
	}

	rm_detail := t.detail.FindRMDetail(seatid)
	if rm_detail != nil {
		rm_detail.ChuCards = append(rm_detail.ChuCards, card)
	}
}

// 胡牌
func (t *Desk) finish(userid string, req *pb.RMFinishReq) {
	msg := new(pb.RMFinishRsp)
	msg1 := new(pb.RMFinishNtf)

	seatid := t.getSeatid(userid)

	if t.ActSeat != seatid {
		msg.Error = pb.NotYourTurn
		t.send2userid(userid, msg)
		return
	}

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)
	if seat == nil || status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if len(seat.Cards) != 14 {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var req_cards []uint32
	req_cards = append(req_cards, req.Card)
	for _, v := range req.Infos {
		req_cards = append(req_cards, v.Cards...)
	}

	if !algo.IsSame(req_cards, seat.Cards) {
		msg.Error = pb.CardsNotSame
		t.send2userid(userid, msg)
		return
	}

	var new []uint32
	found := false
	for _, v := range seat.Cards {
		if !found && v == req.Card {
			found = true
			continue
		}
		new = append(new, v)
	}

	if !found {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	} else {
		seat.Cards = new
	}

	var finish [][]uint32
	for _, v := range req.Infos {
		if len(v.Cards) == 0 {
			continue
		}
		finish = append(finish, v.Cards)
	}

	ret := algo.IsFinish(finish, t.WildCard)
	if ret { //真胡
		seat.SortCards = finish

		msg.Isfinish = true
		t.send2userid(userid, msg)

		msg1.Huseat = seatid
		msg1.Isfinish = true
		msg1.Card = req.Card
		t.broadcast(msg1)

		seat.Winner = true
		seat.Finish = req.Card

		//改变状态
		t.state = int32(pb.STATE_DECLARE)
		t.pushState()

	} else { //诈胡
		seat.SortCards = finish

		msg.Isfinish = false
		t.send2userid(userid, msg)

		msg1.Huseat = seatid
		msg1.Isfinish = false
		t.broadcast(msg1)

		seat.Finish = req.Card

		status.Alive = false
		status.ZhaHu = true
		t.Lose(seatid, true, 0, false, 0)
		// status.ActScore = -80

		// num := int64(80 * t.Game.RM.Bottom)
		// t.setBet(seatid, userid, num, fmt.Sprintf("rm%s房间输分", t.DeskData.Rid))

		t.setNextActSeat()
	}
}

// 声明牌
func (t *Desk) declare(userid string, req *pb.RMDeclareReq) {
	msg := new(pb.RMDeclareRsp)
	msg1 := new(pb.RMDeclareNtf)

	seatid := t.getSeatid(userid)

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)

	if seat == nil || status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if t.state != int32(pb.STATE_DECLARE) {
		msg.Error = pb.OperateError
		t.send2userid(userid, msg)
		return
	}

	if !status.Alive {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var req_cards []uint32
	for _, v := range req.Infos {
		req_cards = append(req_cards, v.Cards...)
	}

	if !algo.IsSame(req_cards, seat.Cards) {
		msg.Error = pb.CardsNotSame
		t.send2userid(userid, msg)
		return
	}

	seat.SortCards = [][]uint32{}
	for _, v := range req.Infos {
		seat.SortCards = append(seat.SortCards, v.Cards)
	}

	seat.Declare = true

	t.send2userid(userid, msg)

	msg1.Seat = seatid
	t.broadcast(msg1)

	if t.allDeclare() {
		winner := t.getWinner()
		t.gameOver(winner)
	}
}

func (t *Desk) allDeclare() bool {
	for k, v := range t.seats {
		if !v.Ready { //跳过观战
			continue
		}
		if v.Winner { //跳过赢家
			continue
		}
		status := t.getStatus(k)
		if status.Alive { //还在玩的玩家
			if !v.Declare {
				return false
			}
		}
	}
	return true
}

// 摆牌
func (t *Desk) sort(userid string, req *pb.RMSortReq) {
	msg := new(pb.RMSortRsp)
	seatid := t.getSeatid(userid)

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)
	role := t.getRole(userid)
	if seat == nil || status == nil || role == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if !status.Alive {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var req_cards []uint32
	for _, v := range req.Infos {
		req_cards = append(req_cards, v.Cards...)
	}

	if !algo.IsSame(req_cards, seat.Cards) {
		msg.Error = pb.CardsNotSame
		msg.Cards = seat.Cards
		t.send2userid(userid, msg)
		return
	}

	seat.SortCards = [][]uint32{}
	for _, v := range req.Infos {
		seat.SortCards = append(seat.SortCards, v.Cards)
	}

	t.send2userid(userid, msg)

	if seatid == t.DeskAct.ActSeat {
		seat.Sort = true //设置sort标记
	}

	if IsDebug && role.Robot {
		msg1 := new(pb.RMSortNtf)
		msg1.Seat = seatid
		msg1.Infos = req.Infos
		t.broadcast(msg1)
	}
}

// 弃牌
func (t *Desk) drop(userid string, req *pb.RMDropReq) {
	msg := new(pb.RMDropRsp)
	msg1 := new(pb.RMDropNtf)

	seatid := t.getSeatid(userid)
	if t.ActSeat != seatid {
		msg.Error = pb.NotYourTurn
		t.send2userid(userid, msg)
		return
	}

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)
	if seat == nil || status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if !status.Alive {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	moNum := 0
	score := 0
	rm_detail := t.detail.FindRMDetail(seatid)
	if rm_detail != nil {
		moNum = len(rm_detail.MoCards)
	}

	if moNum == 0 {
		score = 20
	} else if moNum == 1 {
		score = 40
	} else {
		score = 80
	}

	seat.SortCards = [][]uint32{}
	for _, v := range req.Sort {
		seat.SortCards = append(seat.SortCards, v.Cards)
	}

	status.Alive = false
	status.Drop = true
	t.seatDrop(userid)
	t.Lose(seatid, false, 0, true, score)

	t.send2userid(userid, msg)

	msg1.Seat = seatid
	t.broadcast(msg1)

	if t.ActSeat == seatid {
		t.setNextActSeat()
	}
}

func (t *Desk) dropScore(userid string, req *pb.RMDropScoreReq) {
	msg := new(pb.RMDropScoreRsp)

	seatid := t.getSeatid(userid)
	// if t.ActSeat != seatid {
	// 	msg.Error = pb.NotYourTurn
	// 	t.send2userid(userid, msg)
	// 	return
	// }

	seat := t.getSeat(seatid)
	status := t.getStatus(seatid)
	if seat == nil || status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if !status.Alive {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	moNum := 0
	score := 0
	rm_detail := t.detail.FindRMDetail(seatid)
	if rm_detail != nil {
		moNum = len(rm_detail.MoCards)
	}

	if moNum == 0 {
		score = 20
	} else if moNum == 1 {
		score = 40
	} else {
		score = 80
	}

	msg.Score = uint32(score * int(t.Game.RM.Bottom))
	t.send2userid(userid, msg)
}

// '玩家准备
func (t *Desk) readying(userid string, ready bool) (msg *pb.RMReadyRsp) {
	msg = new(pb.RMReadyRsp)
	if t.isFree() {
		msg.Error = pb.OperateError
		return
	}
	//投票中不能准备
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		if t.DeskPriv.VoteSeat != 0 {
			msg.Error = pb.VotingCantLaunchVote
			return
		}
	}
	//已经开始不能准备
	if t.state != int32(pb.STATE_READY) {
		msg.Error = pb.GameStarted
		return
	}
	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		return
	}
	if user.GetCoin() < int64(t.DeskData.Ante) {
		msg.Error = pb.NotEnoughCoin
		return
	}
	//设置状态
	seat := t.getSeatid(userid)
	if v, ok := t.seats[seat]; ok {
		v.Ready = ready
	}
	msg.Seat = seat
	msg.Ready = ready
	t.broadcast(msg) //广播消息
	if t.allReady() {
		t.gameStart() //开始牌局
	}
	return
}

//.

// '操作验证
func (t *Desk) coinCheck(seat uint32, val int32) pb.ErrCode {
	if t.isFree() {
		return pb.OperateError
	}
	if t.state == int32(pb.STATE_READY) {
		return pb.GameNotStart
	}
	if t.DeskAct.ActSeat != seat {
		glog.Errorf("coinCheck err %d, %d, %d",
			t.DeskAct.ActSeat, seat, val)
		return pb.NotYourTurn
	}
	if (t.DeskAct.ActState & val) != val {
		glog.Errorf("coinCheck err %d, %d, %d",
			t.DeskAct.ActState, seat, val)
		return pb.ErrorOperateValue
	}
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		if !v.Alive {
			return pb.AlreadyFold
		}
	} else {
		return pb.NotReady
	}
	return pb.OK
}

// 刷新sideshow
func (t *Desk) refreshSideshow() {
	seat := t.DeskAct.ActSeat
	prev := t.getPrevActSeat()

	if t.DeskAct.ActSeats[seat].See && t.DeskAct.ActSeats[prev].See {
		t.DeskAct.ActState |= int32(pb.ACT_SIDESHOW)
	}
}

// .轮次详情记录
func (t *Desk) turnDetail(userid string, action string) {
	seat := t.getSeatid(userid)
	tpdetail := t.detail.FindTPDetail(seat)
	if tpdetail == nil {
		return
	}
	tpdeitalturn := tpdetail.FindTPDetailTurn(int(t.ActTimes))
	if tpdeitalturn == nil {
		return
	}
	tpdeitalturn.Operation += action
	tpdeitalturn.Operation += ","
}

// '看牌
func (t *Desk) coinSee(userid string) {
	msg := new(pb.RMCoinSeeRsp)
	seat := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_SEE))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	if t.state != int32(pb.STATE_BET) {
		msg.Error = pb.GameNotStart
		t.send2userid(userid, msg)
		return
	}

	msg.Seat = seat
	msg.Userid = userid
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		v.See = true
		if val, ok := t.seats[seat]; ok {
			msg.Cards = val.Cards
		}
		//操作值中去掉看牌
		// t.DeskHua.ActState ^= int32(pb.ACT_SEE)
		t.send2userid(userid, msg)
		t.refreshSideshow()

	}
	msg2 := new(pb.RMCoinSeeNtf)
	msg2.Seat = seat
	msg2.Userid = userid
	msg2.Actseat = t.DeskAct.ActSeat
	msg2.Actstate = uint32(t.DeskAct.ActState)

	t.broadcast(msg2)

	t.turnDetail(userid, "see")
}

//.

// '弃牌
func (t *Desk) coinFold(userid string) {
	msg := new(pb.RMCoinFoldRsp)
	seat := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_PACK))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }

	msg.Seat = seat
	msg.Userid = userid

	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		v.Alive = false
		v.Pack = true
	}
	//操作值中去掉操作
	// t.DeskHua.ActState ^= int32(pb.ACT_PACK)
	t.send2userid(userid, msg)

	//广播弃牌
	msg2 := new(pb.RMCoinFoldNtf)
	msg2.Seat = seat
	msg2.Userid = userid
	t.broadcast(msg2)

	t.turnDetail(userid, "pack")

	//结束或者继续
	t.setNextActSeat()
}

//.

// '跟注
func (t *Desk) coinCall(userid string) {
	msg := new(pb.RMCoinCallRsp)
	seat := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_CALL))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }

	var num = t.ActAnte
	if t.isSee(seat) {
		num = num * 2 //看牌double
	}

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}

	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		t.send2userid(userid, msg)
		t.coinFold(userid) //金币不足弃牌
		return
	}

	msg.Seat = seat
	msg.Userid = userid
	t.send2userid(userid, msg)
	//验证跟注数量
	// var double int64 = 1
	// if t.isSee(seat) {
	// 	double = 2
	// }
	// glog.Debugf("userid %s, double %d, num %d", userid, double, num)
	// glog.Debugf("userid %s, raise %d, call %d",
	// 	userid, t.DeskHua.ActRaiseNum, t.DeskHua.ActCallNum)
	// if (double * t.DeskHua.ActCallNum) != num {
	// 	msg.Error = pb.CallError
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	//广播消息

	msg2 := new(pb.RMCoinCallNtf)

	msg2.Info = t.setBet(seat, userid, num, fmt.Sprintf("tp%s房间跟注", t.DeskData.Rid))
	msg2.Pot = t.DeskGame.BetNum
	if !t.isSee(seat) {
		msg2.State |= int32(pb.ACT_BLIND)
	} else {
		msg2.State |= int32(pb.ACT_CHAAL)
	}
	t.broadcast(msg2)

	if !t.isSee(seat) {
		t.turnDetail(userid, fmt.Sprintf("blind:%d", num))
	} else {
		t.turnDetail(userid, fmt.Sprintf("chaal:%d", num))
	}

	//结束或者继续
	t.setNextActSeat()
}

//.

// '加注
func (t *Desk) coinRaise(userid string) {
	msg := new(pb.RMCoinRaiseRsp)
	seat := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_RAISE))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }

	t.ActAnte *= 2
	var num = t.ActAnte
	if t.isSee(seat) {
		num = num * 2 //看牌double
	}

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}
	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		t.send2userid(userid, msg)
		t.coinFold(userid)
		return
	}

	msg.Seat = seat
	msg.Userid = userid
	t.send2userid(userid, msg)

	//验证加注数量
	// var double int64 = 1
	// if t.isSee(seat) {
	// 	double = 2
	// }
	// glog.Debugf("userid %s, double %d, num %d", userid, double, num)
	// glog.Debugf("userid %s, raise %d, call %d",
	// 	userid, t.DeskHua.ActRaiseNum, t.DeskHua.ActCallNum)
	// n := double * t.DeskHua.ActRaiseNum
	// i := num / n
	// if ((i % double) != 0) || ((num % n) != 0) {
	// 	msg.Error = pb.RaiseError
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	//广播消息
	msg2 := new(pb.RMCoinRaiseNtf)
	msg2.Info = t.setBet(seat, userid, num, fmt.Sprintf("tp%s房间加注", t.DeskData.Rid))
	msg2.Pot = t.DeskGame.BetNum
	if !t.isSee(seat) {
		msg2.State |= int32(pb.ACT_BLIND)
	} else {
		msg2.State |= int32(pb.ACT_CHAAL)
	}
	t.broadcast(msg2)

	if !t.isSee(seat) {
		t.turnDetail(userid, fmt.Sprintf("blindx2:%d", num))
	} else {
		t.turnDetail(userid, fmt.Sprintf("chaalx2:%d", num))
	}
	//加注底池
	// if double == 1 {
	// 	t.DeskHua.ActRaiseNum += num
	// 	t.DeskHua.ActCallNum = num
	// } else {
	// 	t.DeskHua.ActRaiseNum += (num / 2)
	// 	t.DeskHua.ActCallNum = (num / 2)
	// }
	//结束或者继续
	t.setNextActSeat()
}

// 回复比牌
func (t *Desk) coinReplyBi(userid string, agree bool) {
	if agree { //同意比牌
		msg := new(pb.RMCoinReplyBiRsp)
		t.send2userid(userid, msg)

		msg2 := new(pb.RMCoinBiNtf)

		seat := t.getNextActSeat()
		biseat := t.getSeatid(userid)

		cs1 := t.getHandCards(seat)
		cs2 := t.getHandCards(biseat)
		if algo.HuaCompare(cs1, cs2) {
			//cs1 赢
			msg2.Winseat = seat
			msg2.Loseseat = biseat
		} else {
			//cs2 赢
			msg2.Winseat = biseat
			msg2.Loseseat = seat
		}
		if v, ok := t.DeskAct.ActSeats[msg2.Loseseat]; ok {
			v.Alive = false
			v.Lose = true
		}
		a1 := algo.HuaType(cs1)
		a2 := algo.HuaType(cs2)

		msg2.State |= int32(pb.ACT_SIDESHOW)
		msg2.Seat = seat
		msg2.Biseat = biseat
		msg2.Cards = cs1
		msg2.Bicards = cs2
		msg2.Type = a1
		msg2.Bitype = a2

		t.broadcast(msg2)

		t.turnDetail(userid, "agree")

		t.DeskAct.ActSeat = seat
		t.pauseGame(PAUSE_REASON_1, 3, nil)
		//延时3秒
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		// t.setNextActSeat()
	} else { //拒绝比牌
		msg1 := new(pb.RMCoinReplyBiRsp)
		t.send2userid(userid, msg1)

		next := t.getNextActSeat()

		msg := new(pb.RMCoinReplyBiNtf)
		msg.Seat = next
		msg.Biseat = t.getSeatid(userid)
		msg.Agree = false
		t.broadcast(msg)

		t.turnDetail(userid, "refuse")

		t.DeskAct.ActSeat = next
		t.setNextActSeat()

	}
}

// '比牌
func (t *Desk) coinBi(userid string) {
	msg := new(pb.RMCoinBiRsp)

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}

	seat := t.getSeatid(userid)
	num := t.ActAnte
	if t.isSee(seat) {
		num = num * 2 //看牌double
	}

	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		t.send2userid(userid, msg)
		t.coinFold(userid) //金币不足弃牌
		return
	}

	if t.DeskAct.ActState&int32(pb.ACT_SHOW) == int32(pb.ACT_SHOW) { //show

		// msg.Error = t.coinCheck(biseat, int32(pb.ACT_BI))
		// if msg.Error != pb.OK {
		// 	t.send2userid(userid, msg)
		// 	return
		// }
		biseat := t.getPrevActSeat()
		msg.Seat = seat
		msg.Biseat = biseat
		cs1 := t.getHandCards(seat)
		cs2 := t.getHandCards(biseat)
		//TODO 扣除金币,比牌金额不足怎么处理
		if algo.HuaCompare(cs1, cs2) {
			//cs1 赢
			msg.Winseat = seat
			msg.Loseseat = biseat
		} else {
			//cs2 赢
			msg.Winseat = biseat
			msg.Loseseat = seat
		}
		if v, ok := t.DeskAct.ActSeats[msg.Loseseat]; ok {
			v.Alive = false
		}
		//操作值中去掉操作
		// t.DeskHua.ActState ^= int32(pb.ACT_BI)
		t.send2userid(userid, msg)

		msg3 := new(pb.RMCoinCallNtf)
		msg3.Info = t.setBet(seat, userid, num, fmt.Sprintf("tp%s房间show", t.DeskData.Rid))
		msg3.Pot = t.DeskGame.BetNum
		t.broadcast(msg3)

		a1 := algo.HuaType(cs1)
		a2 := algo.HuaType(cs2)

		msg2 := new(pb.RMCoinBiNtf)
		msg2.State |= int32(pb.ACT_SHOW)
		msg2.Seat = msg.Seat
		msg2.Biseat = msg.Biseat
		msg2.Cards = cs1
		msg2.Bicards = cs2
		msg2.Type = a1
		msg2.Bitype = a2
		msg2.Winseat = msg.Winseat
		msg2.Loseseat = msg.Loseseat

		t.broadcast(msg2)

		t.turnDetail(userid, fmt.Sprintf("show:%d", num))

		t.pauseGame(PAUSE_REASON_1, 3, nil)
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		//结束或者继续
		// t.setNextActSeat()
	} else if t.DeskAct.ActState&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) { //sideshow

		t.send2userid(userid, msg)

		seat := t.getSeatid(userid)

		// num := t.ActAnte * 2
		msg2 := new(pb.RMCoinCallNtf)
		msg2.Info = t.setBet(seat, userid, num, fmt.Sprintf("tp%s房间sideshow", t.DeskData.Rid))
		msg2.Pot = t.DeskGame.BetNum
		t.broadcast(msg2)

		t.turnDetail(userid, fmt.Sprintf("sideshow:%d", num))

		t.setPrevActSeat()
	}

	// curr := t.getNextActSeat()
	// //结束,winner = last
	// if curr == 0 {
	// 	t.gameOver()
	// 	return
	// }
	// //不切换位置
	// t.timer = 0
	// //广播操作状态消息
	// t.pushActState()
}

//.

// '换房间
func (t *Desk) changeDesk(ctx actor.Context) {
	userid := t.getRouter(ctx)
	errcode := t.changeDeskCheck(userid)
	if errcode != pb.OK {
		rsp := new(pb.RMCoinChangeRoomRsp)
		rsp.Error = errcode
		ctx.Respond(rsp)
		return
	}
	t.nnLeave(userid, ctx)
	//匹配房间消息
	msg := new(pb.ChangeDesk)
	msg.Roomid = t.DeskData.Rid
	msg.Rtype = t.DeskData.Rtype
	msg.Gtype = t.DeskData.Gtype
	msg.Ltype = t.DeskData.Ltype
	msg.Dtype = t.DeskData.Dtype
	msg.Userid = userid
	msg.Sender = ctx.Sender() //玩家进程
	nodePid.Request(msg, ctx.Self())
}

func (t *Desk) changeDeskCheck(userid string) pb.ErrCode {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		switch t.state {
		case int32(pb.STATE_READY):
		default:
			seat := t.getSeatid(userid)
			if v, ok := t.seats[seat]; ok {
				if v.Ready {
					return pb.GameStartedCannotLeave
				}
			}
		}
	default:
		return pb.OperateError
	}
	return pb.OK
}

// 局内充值开始
func (t *Desk) chargeInGameBegin(seatids ...uint32) {
	if len(seatids) == 0 {
		return
	}
	if t.state == int32(pb.STATE_CHARGE) {
		return
	}
	t.chargePrevState = t.state
	t.chargingSeats = seatids
	t.state = int32(pb.STATE_CHARGE)
	t.timer = 0

	for _, seatid := range seatids {
		if t.Rtype == int32(pb.ROOM_TYPE1) {
			// 私人房回合结束时充值金额推送
			msg1 := new(pb.PrivChargeAmountNtf)
			msg1.Amount, msg1.GiveAmount = t.getRechargeAmount(seatid)
			t.send2seat(seatid, msg1)
		}

		msg2 := new(pb.ChargeInGameNtf)
		msg2.Seat = seatid
		msg2.Totaltimer = ChargeInGameTime + utils.LocalTime().Unix()
		t.broadcast(msg2)
	}
}

// 局内充值取消
func (t *Desk) chargeInGameCancel(userid string) {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	switch t.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		t.state = t.chargePrevState
		if t.state == int32(pb.STATE_READY) {
			t.privRoundChargingResume(true, userid)
		}
	default:
		t.state = int32(pb.STATE_BET)
	}
}

// 私人房回合结束充值恢复
func (t *Desk) privRoundChargingResume(cancel bool, userid string) {
	seatid := t.getSeatid(userid)
	var chargingSeats []uint32
	if cancel {
		// 踢出充值中玩家
		for _, seat := range t.chargingSeats {
			if seat != seatid {
				chargingSeats = append(chargingSeats, seat)
				continue
			}
			userid := t.getUserid(seat)
			//玩家离开牌桌
			t.notifyNodeLeaveEarly(t.Rid, userid, pb.OK)
			t.notifyGateUserLeft(userid, pb.OK, int32(pb.OK))
			//清除数据
			t.userLeaveDesk(userid)
		}
	} else {
		// 充值成功
		for _, seat := range t.chargingSeats {
			if seatid != seat {
				chargingSeats = append(chargingSeats, seat)
			}
		}
	}
	t.chargingSeats = chargingSeats
	// 还有充值中玩家
	if len(t.chargingSeats) > 0 {
		return
	}
	t.state = t.chargePrevState
	if len(t.seats) > 1 {
		t.gameStart()
	} else {
		// 玩家人数不够，结算退出
		t.pushPrivSettle()
	}
}

// 获取局内充值金额
func (t *Desk) getRechargeAmount(seat uint32) (uint32, uint32) {
	rechargeMap := t.DeskAct.ActRechargeTimes
	configs := t.Game.RM.RoomRecharge
	gives := t.Game.RM.RoomRechargeGive
	if len(configs) <= 0 {
		return 0, 0
	}

	if num, ok := rechargeMap[seat]; ok {
		// rechargeMap[seat] = num + 1
		if len(configs) > num {
			return uint32(configs[num]), uint32(gives[num])
		} else {
			return uint32(configs[len(configs)-1]), uint32(gives[len(gives)-1])
		}
	} else {
		// rechargeMap[seat] = 1
		return uint32(configs[0]), uint32(gives[0])
	}
}

// 充值超时
func (t *Desk) chargeTimeout() {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	switch t.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		t.state = t.chargePrevState
		t.privRoundChargingResume(true, "")
	default:
		glog.Errorf("unkonwn chargeTimeout roomType: %v", t.Rtype)
	}
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
