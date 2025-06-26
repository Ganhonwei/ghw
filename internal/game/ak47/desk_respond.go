package ak47

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
	// 	if vip.EmojiCost > 0 {
	// 		t.sendCurrency(userid, -int64(vip.EmojiCost), int32(pb.LOG_TYPE102), "ak47发表情")
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
	//TODO 已经在房间防止数据覆盖
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
	var i uint32
	if len(idleSeats) > 0 {
		winner, err := utils.ChoiceInt(idleSeats)
		if err == nil {
			i = uint32(winner)
		}
	} else {
		i = uint32(utils.RandInt32N(int32(t.DeskData.Count)) + 1) // 先随机一个位置
		var j uint32
		for j = 1; j <= t.DeskData.Count; j++ {
			if _, ok := t.seats[i]; !ok {
				break
			} else {
				i++
				if i > t.DeskData.Count {
					i = 1
				}
			}
		}
	}
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
	if _, ok := t.roles[user.GetUserid()]; !ok {
		switch t.DeskData.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
		//可以旁观
		default:
			if uint32(len(t.roles)) >= t.DeskData.Count {
				return pb.RoomFull //人数已满
			}
		}
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
	// arg := ctx.Message().(*pb.AK47LeaveReq)
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
	msg := new(pb.AK47LeaveRsp)
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
		if !v.Offline {
			msg := new(pb.LeftDesk)
			msg.Error = err
			msg.Gtype = int32(pb.AK47)
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
		if !v.Offline {
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
	if v, ok := t.roles[userid]; ok {
		//离开状态消息
		msg1 := new(pb.AK47LeaveRsp)
		msg1.Seat = v.Seat
		msg1.Userid = userid
		t.send2userid(userid, msg1)

		//广播
		msg2 := new(pb.AK47LeaveNtf)
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
		msg := new(pb.AK47PushOfflineNtf)
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
	msg := handler.BeAK47DealerMsg(st, num, t.DeskGame.Dealer,
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
	msg := handler.BeAK47DealerMsg(0, num, t.DeskGame.Dealer,
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
	msg := handler.BeAK47DealerMsg(1, num, t.DeskGame.Dealer,
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
	msg := new(pb.AK47Ready2Rsp)

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

//.

// '玩家准备
func (t *Desk) readying(userid string, ready bool) (msg *pb.AK47ReadyRsp) {
	msg = new(pb.AK47ReadyRsp)
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
	// seat := t.DeskAct.ActSeat
	// prev := t.getPrevActSeat()

	// if t.DeskAct.ActSeats[seat].See && t.DeskAct.ActSeats[prev].See {
	// 	t.DeskAct.ActState |= int32(pb.ACT_SIDESHOW)
	// }
	t.DeskAct.ActState = t.LimitShow(t.DeskAct.ActState)
}

func (t *Desk) recordBet(betAmount int64) {
	if _, ok := t.betInfo[betAmount]; ok {
		t.betInfo[betAmount]++
	} else {
		t.betInfo[betAmount] = 1
	}
}

// .轮次详情记录
func (t *Desk) turnDetail(userid string, action string) {
	seat := t.getSeatid(userid)
	ak47detail := t.detail.FindAK47Detail(seat)
	if ak47detail == nil {
		return
	}
	ak47deitalturn := ak47detail.FindAK47DetailTurn(int(t.ActTimes))
	if ak47deitalturn == nil {
		return
	}
	ak47deitalturn.Operation += action
	ak47deitalturn.Operation += ","
}

// '看牌
func (t *Desk) coinSee(userid string) {
	msg := new(pb.AK47CoinSeeRsp)
	seat := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_SEE))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	if t.state != int32(pb.STATE_BET) && t.state != int32(pb.STATE_CHARGE) {
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
			msg.Changecards = val.ChangeCards
			msg.Typ = algo.HuaType(msg.Changecards)
		}
		//操作值中去掉看牌
		// t.DeskHua.ActState ^= int32(pb.ACT_SEE)
		t.send2userid(userid, msg)
		t.refreshSideshow()

	}
	msg2 := new(pb.AK47CoinSeeNtf)
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
	t.chargeInGameCancel()

	msg := new(pb.AK47CoinFoldRsp)
	seat := t.getSeatid(userid)

	if !t.leaveCheck(seat) {
		msg.Error = pb.WinnerExitLater
		t.send2userid(userid, msg)
		return
	}
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
		t.Lose(seat)
	}
	//操作值中去掉操作
	// t.DeskHua.ActState ^= int32(pb.ACT_PACK)
	t.send2userid(userid, msg)

	//广播弃牌
	msg2 := new(pb.AK47CoinFoldNtf)
	msg2.Seat = seat
	msg2.Userid = userid
	t.broadcast(msg2)

	t.turnDetail(userid, "pack")

	//结束或者继续
	t.setNextActSeat()
}

// chargeInGameRobot 机器人假装局内充值，条件：
// 人机牌型大于等于同花，
// 人机本轮判定的行为是下注或加注，
// 有2%的概率触发假充值。
// 充值时间20-40秒随机
// 触发后的人机本局必会离开桌子（防止穿帮）
// return isCharging
func (t *Desk) chargeInGameRobot(userid string) (charging bool) {
	return false
}
func (t *Desk) chargeInGameRobot0(userid string) (charging bool) {
	// 匹配
	if t.Rtype != int32(pb.ROOM_TYPE0) {
		return
	}
	role, ok := t.roles[userid]
	if !ok || !role.User.Robot {
		return
	}
	seatid := role.Seat
	if !utils.RandWan(200) {
		return
	}
	// 触发过了
	if 0 < t.DeskAct.ActRechargeTimes[seatid] {
		return
	}
	// 牌型大于等于同花
	cs := t.getHandChangeCards(seatid)
	if algo.TongHua > algo.HuaType(cs) {
		return
	}

	t.robotChargingTime = utils.RandMN(20, 40)
	glog.Infof("robot charging in game start: %s, %v", userid, t.robotChargingTime)
	t.chargeInGameBegin(seatid)
	return true
}

// 局内充值开始
func (t *Desk) chargeInGameBegin(seatid uint32) {
	if t.state == int32(pb.STATE_CHARGE) {
		return
	}
	t.state = int32(pb.STATE_CHARGE)
	t.timer = 0
	t.chargingSeat = seatid

	msg2 := new(pb.ChargeInGameNtf)
	msg2.Seat = seatid
	msg2.Totaltimer = ChargeInGameTime + utils.LocalTime().Unix()
	t.broadcast(msg2)
}

// 局内充值取消
func (t *Desk) chargeInGameCancel() {
	if t.state != int32(pb.STATE_CHARGE) {
		return
	}
	t.state = int32(pb.STATE_BET)
	t.chargingSeat = 0
}

// 向上整百值
func (t *Desk) roundUp(n int64) int64 {
	remainder := n % 10000
	if remainder == 0 {
		return n
	}
	return n + 10000 - remainder
}

// 获取局内充值信息
func (t *Desk) getChargeInfo(seatid uint32) *pb.AK47ChargeInGameInfo {
	seat := t.getSeat(seatid)
	if seat == nil {
		return nil
	}
	info := &pb.AK47ChargeInGameInfo{}
	info.WinRate = int32(seat.WinRate * 10000)
	info.ChargeTime = int32(ChargeInGameTime - t.timer)
	info.ChargeWin = int32(t.getEstimateWin2(seatid))

	return info
}

func (t *Desk) getEstimateWin2(seat uint32) int64 {
	count := 0
	for k, v := range t.seats {
		if !v.Ready {
			continue
		}
		status := t.getStatus(k)
		if status == nil {
			continue
		}
		if !status.Alive {
			continue
		}
		count++
	}
	return t.DeskGame.BetNum + t.ActAnte*int64(count)
}

// '跟注
func (t *Desk) coinCall(userid string) {
	t.chargeInGameCancel()

	msg := new(pb.AK47CoinCallRsp)
	seatid := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_CALL))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	if !t.CanOper(seatid) {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if t.ActSeat != seatid {
		glog.Errorf("NotYourTurn,userid:%s,seatid:%d", userid, seatid)
		msg.Error = pb.NotYourTurn
		t.send2userid(userid, msg)
		return
	}

	status := t.getStatus(seatid)
	if status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var num = t.ActAnte
	if t.isSee(seatid) {
		num = num * 2 //看牌double
		status.Chaal = true
	}

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}

	// 机器人假装局内充值
	if t.chargeInGameRobot(userid) {
		return
	}

	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
		msg.ChargeInfo = t.getChargeInfo(t.DeskAct.ActSeat)
		t.send2userid(userid, msg)
		t.chargeInGameBegin(seatid)
		// t.coinFold(userid) //金币不足弃牌
		return
	}

	msg.Seat = seatid
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

	msg2 := new(pb.AK47CoinCallNtf)

	msg2.Info = t.setBet(seatid, userid, num, fmt.Sprintf("ak47%s房间跟注", t.DeskData.Rid))
	t.recordBet(num)
	msg2.Pot = t.DeskGame.BetNum
	if !t.isSee(seatid) {
		msg2.State |= int32(pb.ACT_BLIND)
	} else {
		msg2.State |= int32(pb.ACT_CHAAL)
	}
	t.broadcast(msg2)

	if !t.isSee(seatid) {
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
	t.chargeInGameCancel()

	msg := new(pb.AK47CoinRaiseRsp)
	seatid := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_RAISE))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	if !t.CanOper(seatid) {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	status := t.getStatus(seatid)
	if status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	var num = t.ActAnte
	num *= 2
	if t.isSee(seatid) {
		num = num * 2 //看牌double
		status.Chaal = true
	}

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}

	// 机器人假装局内充值
	if t.chargeInGameRobot(userid) {
		return
	}

	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
		msg.ChargeInfo = t.getChargeInfo(t.DeskAct.ActSeat)
		t.send2userid(userid, msg)
		t.chargeInGameBegin(seatid)
		// t.coinFold(userid)
		return
	}

	t.ActAnte *= 2

	msg.Seat = seatid
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
	msg2 := new(pb.AK47CoinRaiseNtf)
	msg2.Info = t.setBet(seatid, userid, num, fmt.Sprintf("ak47%s房间加注", t.DeskData.Rid))
	t.recordBet(num)
	msg2.Pot = t.DeskGame.BetNum
	if !t.isSee(seatid) {
		msg2.State |= int32(pb.ACT_BLIND)
	} else {
		msg2.State |= int32(pb.ACT_CHAAL)
	}
	t.broadcast(msg2)

	if !t.isSee(seatid) {
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

// 污染检查
func (t *Desk) DirtyCheck(seat1, seat2 uint32) {
	pair := [2]uint32{seat1, seat2}
	t.cmpRecord = append(t.cmpRecord, pair)

	for _, v := range t.cmpRecord {
		t._DirtyCheck(v[0], v[1])
	}

	for i := len(t.cmpRecord) - 1; i >= 0; i-- {
		t._DirtyCheck(t.cmpRecord[i][0], t.cmpRecord[i][1])
	}
}

func (t *Desk) _DirtyCheck(seat1, seat2 uint32) {
	status1 := t.getStatus(seat1)
	status2 := t.getStatus(seat2)

	if status1 == nil || status2 == nil {
		return
	}

	if status1.Dirty {
		status2.Dirty = true
	} else if status2.Dirty {
		status1.Dirty = true
	}
}

// 获取没有污染的人机座位
func (t *Desk) GetNonDirtySeat() []uint32 {
	var seats []uint32

	for k, _ := range t.seats {
		status := t.getStatus(k)
		if status == nil {
			continue
		}
		if status.Dirty {
			continue
		}
		seats = append(seats, k)
	}
	return seats
}

// 策略局换牌
func (t *Desk) StrategyShowCard() bool {
	if !t.isStrategy {
		return false
	}

	seats := t.GetNonDirtySeat()
	if len(seats) == 0 {
		return false
	}

	var change bool

	var seata, seatb *data.DeskSeat
	for _, v := range seats {
		seat := t.getSeat(v)
		status := t.getStatus(v)
		if seat == nil || status == nil {
			continue
		}

		if seat.Identity == "a" && status.Pack {
			seata = seat
			// t.DeskGame.Cards = append(t.DeskGame.Cards, seat.Cards...)
			for _, v := range seat.Cards {
				if algo.AK47IsLaizi(v) {
					t.DeskGame.LaiCards = append(t.DeskGame.LaiCards, v)
				} else {
					t.DeskGame.Cards = append(t.DeskGame.Cards, v)
				}
			}
		}

		if seat.Identity == "b" && status.Pack {
			seatb = seat
			// t.DeskGame.Cards = append(t.DeskGame.Cards, seat.Cards...)
			for _, v := range seat.Cards {
				if algo.AK47IsLaizi(v) {
					t.DeskGame.LaiCards = append(t.DeskGame.LaiCards, v)
				} else {
					t.DeskGame.Cards = append(t.DeskGame.Cards, v)
				}
			}
		}
	}

	if seata != nil {
		var choices []utils.Choice
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ := utils.WeightedChoice(choices)
		wildCardsNum := ch.Item.(int)

		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
		ch, _ = utils.WeightedChoice(choices)
		cardtype := ch.Item.(uint32)

		seata.ShowCards, seata.ShowChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
		change = true
	}

	if seatb != nil {
		var choices []utils.Choice
		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 30, Item: 0})
		choices = append(choices, utils.Choice{Weight: 70, Item: 1})
		ch, _ := utils.WeightedChoice(choices)
		wildCardsNum := ch.Item.(int)

		choices = []utils.Choice{}
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHuaShun})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
		ch, _ = utils.WeightedChoice(choices)
		cardtype := ch.Item.(uint32)

		seatb.ShowCards, seatb.ShowChangeCards, t.DeskGame.Cards = t.getCard2(t.DeskGame.Cards, wildCardsNum, cardtype)
		change = true
	}

	if seata != nil && seatb != nil {
		if algo.AK47Compare(seatb.ShowCards, seata.ShowCards, seatb.ShowChangeCards, seata.ShowChangeCards) {
			seata.ShowCards, seatb.ShowCards = seatb.ShowCards, seata.ShowCards
			seata.ShowChangeCards, seatb.ShowChangeCards = seatb.ShowChangeCards, seata.ShowChangeCards
		}
	}

	return change
}

// 回复比牌
func (t *Desk) coinReplyBi(userid string, agree bool) {
	if agree { //同意比牌
		msg := new(pb.AK47CoinReplyBiRsp)
		t.send2userid(userid, msg)

		msg2 := new(pb.AK47CoinBiNtf)

		seat := t.getNextActSeat()
		biseat := t.getSeatid(userid)

		cs1 := t.getHandCards(seat)
		cs2 := t.getHandCards(biseat)
		cs3 := t.getHandChangeCards(seat)
		cs4 := t.getHandChangeCards(biseat)

		if algo.AK47Compare(cs1, cs2, cs3, cs4) {
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
			t.Lose(msg2.Loseseat)
		}
		a1 := algo.HuaType(cs3)
		a2 := algo.HuaType(cs4)

		msg2.State |= int32(pb.ACT_SIDESHOW)
		msg2.Seat = seat
		msg2.Biseat = biseat
		msg2.Cards = cs1
		msg2.Bicards = cs2
		msg2.Changecards = cs3
		msg2.Bichangecards = cs4
		msg2.Type = a1
		msg2.Bitype = a2
		msg2.IsShow = false

		msg4 := new(pb.AK47CoinBiNtf)
		utils.Clone(msg4, msg2)
		msg4.Cards = []uint32{}
		msg4.Bicards = []uint32{}
		msg4.Changecards = []uint32{}
		msg4.Bichangecards = []uint32{}

		seats := []uint32{msg2.Seat, msg2.Biseat}
		others := t.GetOtherSeats(seats)

		t.broadcast4(seats, msg2)
		t.broadcast4(others, msg4)

		t.turnDetail(userid, "agree")

		t.DirtyCheck(seat, biseat)

		t.DeskAct.ActSeat = seat
		t.pauseGame(PAUSE_REASON_1, 3, nil)
		//延时3秒
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		// t.setNextActSeat()
	} else { //拒绝比牌
		msg1 := new(pb.AK47CoinReplyBiRsp)
		t.send2userid(userid, msg1)

		next := t.getNextActSeat()

		msg := new(pb.AK47CoinReplyBiNtf)
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
	t.chargeInGameCancel()

	msg := new(pb.AK47CoinBiRsp)

	user := t.getPlayer(userid)
	if user == nil {
		msg.Error = pb.NotInRoom
		t.send2userid(userid, msg)
		return
	}

	seatid := t.getSeatid(userid)
	if !t.CanOper(seatid) {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	num := t.ActAnte
	if t.isSee(seatid) {
		num = num * 2 //看牌double
	}

	if user.GetScore() < num {
		msg.Error = pb.NotEnoughCoin
		msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
		msg.ChargeInfo = t.getChargeInfo(t.DeskAct.ActSeat)
		t.send2userid(userid, msg)
		t.chargeInGameBegin(seatid)
		// t.coinFold(userid) //金币不足弃牌
		return
	}

	if t.DeskAct.ActState&int32(pb.ACT_SHOW) == int32(pb.ACT_SHOW) { //show

		// msg.Error = t.coinCheck(biseat, int32(pb.ACT_BI))
		// if msg.Error != pb.OK {
		// 	t.send2userid(userid, msg)
		// 	return
		// }
		biseat := t.getPrevActSeat()
		msg.Seat = seatid
		msg.Biseat = biseat
		cs1 := t.getHandCards(seatid)
		cs2 := t.getHandCards(biseat)
		cs3 := t.getHandChangeCards(seatid)
		cs4 := t.getHandChangeCards(biseat)
		//TODO 扣除金币,比牌金额不足怎么处理
		if algo.AK47Compare(cs1, cs2, cs3, cs4) {
			//cs1 赢
			msg.Winseat = seatid
			msg.Loseseat = biseat
		} else {
			//cs2 赢
			msg.Winseat = biseat
			msg.Loseseat = seatid
		}
		if v, ok := t.DeskAct.ActSeats[msg.Loseseat]; ok {
			v.Alive = false
			t.Lose(msg.Loseseat)
		}
		//操作值中去掉操作
		// t.DeskHua.ActState ^= int32(pb.ACT_BI)
		t.send2userid(userid, msg)

		msg3 := new(pb.AK47CoinCallNtf)
		msg3.Info = t.setBet(seatid, userid, num, fmt.Sprintf("ak47%s房间show", t.DeskData.Rid))
		t.recordBet(num)
		msg3.Pot = t.DeskGame.BetNum
		t.broadcast(msg3)

		a1 := algo.HuaType(cs3)
		a2 := algo.HuaType(cs4)

		msg2 := new(pb.AK47CoinBiNtf)
		msg2.State |= int32(pb.ACT_SHOW)
		msg2.Seat = msg.Seat
		msg2.Biseat = msg.Biseat
		msg2.Cards = cs1
		msg2.Bicards = cs2
		msg2.Changecards = cs3
		msg2.Bichangecards = cs4
		msg2.Type = a1
		msg2.Bitype = a2
		msg2.Winseat = msg.Winseat
		msg2.Loseseat = msg.Loseseat
		msg2.IsShow = true

		msg4 := new(pb.AK47CoinBiNtf)
		utils.Clone(msg4, msg2)
		msg4.Cards = []uint32{}
		msg4.Bicards = []uint32{}
		msg4.Changecards = []uint32{}
		msg4.Bichangecards = []uint32{}

		seats := []uint32{msg2.Seat, msg2.Biseat}
		others := t.GetOtherSeats(seats)

		t.broadcast4(seats, msg2)
		t.broadcast4(others, msg4)

		t.turnDetail(userid, fmt.Sprintf("show:%d", num))

		t.DirtyCheck(seatid, biseat)

		t.pauseGame(PAUSE_REASON_1, 3, nil)
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		//结束或者继续
		// t.setNextActSeat()
	} else if t.DeskAct.ActState&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) { //sideshow

		t.send2userid(userid, msg)

		seat := t.getSeatid(userid)

		// num := t.ActAnte * 2
		msg2 := new(pb.AK47CoinCallNtf)
		msg2.Info = t.setBet(seat, userid, num, fmt.Sprintf("ak47%s房间sideshow", t.DeskData.Rid))
		t.recordBet(num)
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
		rsp := new(pb.AK47CoinChangeRoomRsp)
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

// 获取局内充值金额
func (t *Desk) getRechargeAmount(seat uint32) (uint32, uint32) {
	if t.isStrategy && t.chargeAmount != 0 {
		return t.chargeAmount, uint32(float64(t.chargeAmount) * 0.05)
	} else {
		rechargeMap := t.DeskAct.ActRechargeTimes
		configs := t.Game.AK47.RoomRecharge
		gives := t.Game.AK47.RoomRechargeGive
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
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
