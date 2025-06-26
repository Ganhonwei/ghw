package hua

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"

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
	// 		t.sendCurrency(userid, -int64(vip.EmojiCost), int32(pb.LOG_TYPE102), "tp发表情")
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

	// 人机初始携带金额
	if user.Robot {
		zlog.Infof("%s robot enter desk: carry=%d", t.GameId, user.GetScore())
	}
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
				// p.Seat = i
				// t.seats[i] = &data.DeskSeat{
				// 	Userid: p.User.GetUserid(),
				// }
				// if t.IsGaming() {
				// 	t.seats[i].Watch = true
				// }
				break
			} else {
				i++
				if i > t.DeskData.Count {
					i = 1
				}
			}
		}
	}
	glog.Debugf("user:%s ,seat:%d", p.Userid, i)
	p.Seat = i
	t.seats[i] = &data.DeskSeat{
		Userid: p.User.GetUserid(),
	}
	// // 私人房自动准备
	// if t.Rtype == int32(pb.ROOM_TYPE1) {
	// 	t.seats[i].Ready = true
	// }
	if t.IsGaming() {
		t.seats[i].Watch = true
	} else {
		// 自动准备
		t.seats[i].Ready = true
	}
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
			return pb.GameStarted
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
	arg := ctx.Message().(*pb.JHLeaveReq)
	//离线
	defer t.offlineDesk(userid)
	//seat := t.getSeat(userid)
	errcode := t.leave(userid)
	t.notifyNodeLeaveEarly(t.Rid, userid, errcode)
	t.notifyGateUserLeft(userid, errcode, arg.Reason)
	if errcode == pb.OK || errcode == pb.LeaveEarly {
		//清除数据
		t.userLeaveDesk(userid)
		return
	}
	msg := new(pb.JHLeaveRsp)
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
			msg.Gtype = int32(pb.HUA)
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
		if t.IsGaming() { //游戏进行中
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
		} else { //游戏没开始
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
		msg1 := new(pb.JHLeaveRsp)
		msg1.Seat = v.Seat
		msg1.Userid = userid
		t.send2userid(userid, msg1)

		//广播
		msg2 := new(pb.JHLeaveNtf)
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
		msg := new(pb.JHPushOfflineNtf)
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
	msg := handler.BeJHDealerMsg(st, num, t.DeskGame.Dealer,
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
	msg := handler.BeJHDealerMsg(0, num, t.DeskGame.Dealer,
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
	msg := handler.BeJHDealerMsg(1, num, t.DeskGame.Dealer,
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
	msg := new(pb.JHReady2Rsp)

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

	if t.Rtype == int32(pb.ROOM_TYPE1) {
		// 离线玩家无法发送 ready2,下回合开始游戏卡住
		if t.state == int32(pb.STATE_DEALING) {
			var offlines []uint32
			var ready2s int
			for seat, v := range t.seats {
				role := t.roles[v.Userid]
				if role.Offline {
					offlines = append(offlines, seat)
				}
				if v.Ready && v.Ready2 { //跳过未参与玩家
					ready2s++
				}
			}
			// 离线玩家自动准备
			if len(offlines)+ready2s == len(t.seats) {
				for _, offlineSeat := range offlines {
					t.seats[offlineSeat].Ready = true
					t.seats[offlineSeat].Ready2 = true
				}
			}
		}
	}

	if t.state == int32(pb.STATE_DEALING) && t.allReady2() {
		t.state = int32(pb.STATE_BET) //切换状态为下注状态
		t.pushState()
		t.initAct()
	}
}

//.

// '玩家准备
func (t *Desk) readying(userid string, ready bool) (msg *pb.JHReadyRsp) {
	msg = new(pb.JHReadyRsp)
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
	// 	t.DeskAct.ActState = t.LimitShow(t.DeskAct.ActState)
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
	msg := new(pb.JHCoinSeeRsp)
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
			msg.Typ = algo.HuaType(msg.Cards)
		}
		//操作值中去掉看牌
		// t.DeskHua.ActState ^= int32(pb.ACT_SEE)
		t.send2userid(userid, msg)
		t.refreshSideshow()

	}
	msg2 := new(pb.JHCoinSeeNtf)
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
	t.chargeInGameCancel(userid)

	msg := new(pb.JHCoinFoldRsp)
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

	status := t.getStatus(seat)
	if status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	if !status.Alive || status.Pack {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}

	// 记录玩家操作
	var act int32
	act |= int32(pb.ACT_PACK)
	if !t.isSee(seat) {
		act |= int32(pb.ACT_BLIND)
	}
	status.ActHistory = append(status.ActHistory, act)

	status.Alive = false
	status.Pack = true
	t.Lose(seat)

	// if v, ok := t.DeskAct.ActSeats[seat]; ok {
	// 	v.Alive = false
	// 	v.Pack = true
	// 	t.Lose(seat)
	// }

	msg.Seat = seat
	msg.Userid = userid
	//操作值中去掉操作
	// t.DeskHua.ActState ^= int32(pb.ACT_PACK)
	t.send2userid(userid, msg)

	//广播弃牌
	msg2 := new(pb.JHCoinFoldNtf)
	msg2.Seat = seat
	msg2.Userid = userid
	t.broadcast(msg2)

	t.turnDetail(userid, "pack")

	// 只有2人时增加剩余人机看牌概率
	t.addRobotSeeRateWhenJust2Player()
	// 判断人机是否离开
	t.robotLeaveWhenFinish(seat, false)

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
	hand := t.getHandCards(seatid)
	if algo.TongHua > algo.HuaType(hand) {
		return
	}

	t.robotChargingTime = utils.RandMN(20, 40)
	glog.Infof("robot charging in game start: %s, %v", userid, t.robotChargingTime)
	t.chargeInGameBegin(seatid)
	return true
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
		// 私人房
		if t.Rtype == int32(pb.ROOM_TYPE1) &&
			t.chargePrevState != int32(pb.STATE_BET) {
			// 私人房回合结束时充值金额推送
			msg1 := new(pb.PrivChargeAmountNtf)
			msg1.Amount, msg1.GiveAmount = t.getRechargeAmount(seatid)
			c := t.getChargeInfo(seatid)
			msg1.ChargeInfo = &pb.PrivChargeInGameInfo{
				WinRate:    c.WinRate,
				ChargeWin:  c.ChargeWin,
				ChargeTime: c.ChargeTime,
			}
			t.send2seat(seatid, msg1)
		}

		// 局内充值记录
		if !t.isRobot(seatid) {
			amount, _ := t.getRechargeAmount(seatid)
			t.detail.ChargeInGame = append(t.detail.ChargeInGame, [4]int32{int32(seatid), int32(amount), 0, 0})
		} else {
			// 人机计算充值金额
			seat := t.getSeat(seatid)
			player := t.getPlayer(seat.Userid)
			var chargeAmount = t.getRobotChargeInGameAmount(seatid, seat)
			player.Coin += int64(chargeAmount)

			chargeDelay := t.getRobotExpressRecord().RobotChargeInGameDelay
			t.robotChargingTime = utils.RandMN(int(chargeDelay[0]), int(chargeDelay[1])) / 1000
			glog.Infof("%s robot charging in game start: %s, time=%d, carry=%d, charge=%d", t.GameId, seat.Userid, t.robotChargingTime, player.GetScore()-int64(chargeAmount), chargeAmount)
		}
		msg2 := new(pb.ChargeInGameNtf)
		msg2.Seat = seatid
		msg2.Totaltimer = ChargeInGameTime + utils.LocalTime().Unix()
		t.broadcast(msg2)
		t.SetTpUserChargeInGameNumOfTrigger(seatid)
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

	seatid := t.getSeatid(userid)
	zlog.Infof("局内充值取消: %s, %s, %d, %#v", t.GameId, userid, t.model, t.chargingAmounts[seatid])
	delete(t.chargingAmounts, seatid)
}

// 私人房回合结束充值恢复
// cancel && userid == "" , 充值超时
func (t *Desk) privRoundChargingResume(cancel bool, userid string) {
	seatid := t.getSeatid(userid)
	var chargingSeats []uint32
	if cancel {
		// 踢出充值中玩家
		for _, seat := range t.chargingSeats {
			if userid != "" && seat != seatid {
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

// 向上整百值
func (t *Desk) roundUp(n int64) int64 {
	remainder := n % 10000
	if remainder == 0 {
		return n
	}
	return n + 10000 - remainder
}

// 获取局内充值信息
func (t *Desk) getChargeInfo(seatid uint32) *pb.JHChargeInGameInfo {
	seat := t.getSeat(seatid)
	if seat == nil {
		return nil
	}
	info := &pb.JHChargeInGameInfo{}
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
	t.chargeInGameCancel(userid)

	msg := new(pb.JHCoinCallRsp)
	seatid := t.getSeatid(userid)
	// msg.Error = t.coinCheck(seat, int32(pb.ACT_CALL))
	// if msg.Error != pb.OK {
	// 	t.send2userid(userid, msg)
	// 	return
	// }
	if !t.CanOper(seatid) {
		glog.Errorf("no oper,userid:%s,seatid:%d", userid, seatid)
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

	// 记录玩家操作
	var act int32
	act |= int32(pb.ACT_CHAAL)
	if !t.isSee(seatid) {
		act |= int32(pb.ACT_BLIND)
	}
	status.ActHistory = append(status.ActHistory, act)

	// 机器人假装充值
	if t.chargeInGameRobot(userid) {
		return
	}

	// 私人房娱乐模式不判断
	if t.Rtype != int32(pb.ROOM_TYPE1) || t.Gmode != 1 {
		if user.GetScore() < num {
			msg.Error = pb.NotEnoughCoin
			msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
			msg.StrategyType = int32(t.strategyType)
			msg.ChargeInfo = t.getChargeInfo(seatid)
			t.send2userid(userid, msg)
			t.chargeInGameBegin(seatid)
			// t.coinFold(userid) //金币不足弃牌
			return
		}
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

	msg2 := new(pb.JHCoinCallNtf)

	msg2.Info = t.setBet(seatid, userid, num, fmt.Sprintf("tp%s房间跟注", t.DeskData.Rid))
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

		// 人机在其他回合看牌
		t.robotSeeingInOther(seatid)
	}

	//结束或者继续
	t.setNextActSeat()
}

//.

// '加注
func (t *Desk) coinRaise(userid string) {
	t.chargeInGameCancel(userid)

	msg := new(pb.JHCoinRaiseRsp)
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

	if t.ActSeat != seatid {
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
	// 记录玩家操作
	var act int32
	act |= int32(pb.ACT_RAISE)
	if !t.isSee(seatid) {
		act |= int32(pb.ACT_BLIND)
	}
	status.ActHistory = append(status.ActHistory, act)

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

	// 机器人假装充值
	if t.chargeInGameRobot(userid) {
		return
	}

	// 私人房娱乐模式不判断
	if t.Rtype != int32(pb.ROOM_TYPE1) || t.Gmode != 1 {
		if user.GetScore() < num {
			msg.Error = pb.NotEnoughCoin
			msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
			msg.StrategyType = int32(t.strategyType)
			msg.ChargeInfo = t.getChargeInfo(seatid)
			t.send2userid(userid, msg)
			t.chargeInGameBegin(seatid)
			// t.coinFold(userid)
			return
		}
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
	msg2 := new(pb.JHCoinRaiseNtf)
	msg2.Info = t.setBet(seatid, userid, num, fmt.Sprintf("tp%s房间加注", t.DeskData.Rid))
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

		// 人机在其他回合看牌
		t.robotSeeingInOther(seatid)
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

// 随机污染人机
func (t *Desk) RandomDirty() {
	for k, _ := range t.seats {
		status := t.getStatus(k)
		if status == nil {
			continue
		}
		if status.Dirty {
			continue
		}
		if utils.RandWan(5000) { //测试不污染，100%换亮牌
			status.Dirty = true
		}
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

// 检查玩家牌型是否满足条件
// 0：不换 1：小换大 2：大换小
func (t *Desk) CheckPlayerCardType(winner uint32) int {
	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return 0
	}

	seatid := t.getSeatid(userid)
	cards := t.getHandCards(seatid)
	if len(cards) != 3 {
		return 0
	}

	typ := algo.HuaType(cards)
	if winner == seatid { //玩家赢，小换大
		if typ == algo.GaoPai || typ == algo.DuiZi { //对子或单牌
			return 1
		}
	} else { //玩家输，大换小
		if typ == algo.GaoPai { //单牌
			rank := algo.GetGaoPaiRank(cards)
			if rank == algo.Ace || rank >= algo.Ten { //单牌牌值大于等于10
				return 2
			}
		} else {
			return 2
		}
	}
	return 0
}

// 判断亮牌是否启用
func (t *Desk) IsShowCardEnable() bool {
	t.RandomDirty()

	if t.DeskGame.BetNum >= int64(t.Game.TP.Pool_Limit) { //下注筹码达到上限自动开牌不换
		return false
	}

	// if t.DeskGame.BetNum >= 4 { //下注大于等于4轮不换
	// 	return false
	// }

	if len(t.GetNonDirtySeat()) == 0 { //没有污染的人机为0不换
		return false
	}

	r, _ := t.roleCountNum() //两人以上真人局，不换
	if r >= 2 {
		return false
	}

	return true
}

// 升序排序座位(牌越来越大)
func (t *Desk) AscSortSeat(seats []uint32) {
	sort.Slice(seats, func(i, j int) bool {
		cardsi := t.getHandCards(seats[i])
		cardsj := t.getHandCards(seats[j])
		return !algo.HuaCompare(cardsi, cardsj)
	})
}

// 降序排序座位(牌越来越小)
func (t *Desk) DescSortSeat(seats []uint32) {
	sort.Slice(seats, func(i, j int) bool {
		cardsi := t.getHandCards(seats[i])
		cardsj := t.getHandCards(seats[j])
		return algo.HuaCompare(cardsi, cardsj)
	})
}

// 升序
func (t *Desk) AscSortCards(cards [][]uint32) {
	sort.Slice(cards, func(i, j int) bool {
		card1 := cards[i]
		card2 := cards[j]
		return !algo.HuaCompare(card1, card2)
	})
}

// 降序
func (t *Desk) DescSortCards(cards [][]uint32) {
	sort.Slice(cards, func(i, j int) bool {
		card1 := cards[i]
		card2 := cards[j]
		return algo.HuaCompare(card1, card2)
	})
}

// 亮牌
func (t *Desk) ShowCard(winner uint32) bool {
	// if t.isStrategy {
	// 	return false
	// }
	// if t.changeCardType == 1 || t.changeCardType == 2 {
	// 	return false
	// }

	if !t.IsShowCardEnable() {
		return false
	}

	ret := t.CheckPlayerCardType(winner)
	if ret == 0 {
		return false
	}

	seats := t.GetNonDirtySeat()
	for _, v := range seats {
		seat := t.getSeat(v)
		if seat == nil {
			continue
		}
		//将要换的手牌添加回牌堆
		t.DeskGame.Cards = append(t.DeskGame.Cards, seat.Cards...)
	}

	userid := t.GetOnlyOnePlayer()
	if userid == "" {
		return false
	}

	seatid := t.getSeatid(userid)
	cards := t.getHandCards(seatid)
	if len(cards) != 3 {
		return false
	}

	if ret == 1 { //小换大
		t.AscSortSeat(seats)
		showCardsSlice := [][]uint32{}
		for k, _ := range seats {
			_ = k
			var bigger []uint32
			bigger, t.DeskGame.Cards = algo.GetBiggerCard(cards, t.DeskGame.Cards)
			if len(bigger) != 0 {
				showCardsSlice = append(showCardsSlice, bigger)
			}
		}
		if len(seats) != len(showCardsSlice) {
			return false
		}
		t.AscSortCards(showCardsSlice)

		for k, v := range seats {
			seat := t.getSeat(v)
			seat.ShowCards = showCardsSlice[k]
		}

	} else if ret == 2 { //大换小
		t.DescSortSeat(seats)
		showCardsSlice := [][]uint32{}
		for k, _ := range seats {
			_ = k
			var smaller []uint32
			smaller, t.DeskGame.Cards = algo.GetSmallerCard(cards, t.DeskGame.Cards)
			if len(smaller) != 0 {
				showCardsSlice = append(showCardsSlice, smaller)
			}
		}
		if len(seats) != len(showCardsSlice) {
			return false
		}
		t.DescSortCards(showCardsSlice)

		for k, v := range seats {
			seat := t.getSeat(v)
			seat.ShowCards = showCardsSlice[k]
		}

	}

	return true
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
			t.DeskGame.Cards = append(t.DeskGame.Cards, seat.Cards...)
		}

		if seat.Identity == "b" && status.Pack {
			seatb = seat
			t.DeskGame.Cards = append(t.DeskGame.Cards, seat.Cards...)
		}
	}

	if seata != nil {
		var choices []utils.Choice
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
		ch, _ := utils.WeightedChoice(choices)
		cardtype := ch.Item.(uint32)
		seata.ShowCards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
		change = true
	}

	if seatb != nil {
		var choices []utils.Choice
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.ShunZi})
		choices = append(choices, utils.Choice{Weight: 50, Item: algo.TongHua})
		ch, _ := utils.WeightedChoice(choices)
		cardtype := ch.Item.(uint32)
		seatb.ShowCards, t.DeskGame.Cards = algo.GetCard(cardtype, t.DeskGame.Cards)
		change = true
	}

	if seata != nil && seatb != nil {
		if algo.HuaCompare(seatb.ShowCards, seata.ShowCards) {
			seata.ShowCards, seatb.ShowCards = seatb.ShowCards, seata.ShowCards
		}
	}

	return change
}

// 回复比牌
func (t *Desk) coinReplyBi(userid string, agree bool) {
	if agree { //同意比牌
		msg := new(pb.JHCoinReplyBiRsp)
		t.send2userid(userid, msg)

		msg2 := new(pb.JHCoinBiNtf)

		seat := t.getNextActSeat()

		msg3 := new(pb.JHCoinReplyBiNtf)
		msg3.Seat = seat
		msg3.Biseat = t.getSeatid(userid)
		msg3.Agree = true
		t.broadcast(msg3)

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
			t.Lose(msg2.Loseseat)
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
		msg2.IsShow = false

		msg4 := new(pb.JHCoinBiNtf)
		utils.Clone(msg4, msg2)
		msg4.Cards = []uint32{}
		msg4.Bicards = []uint32{}

		seats := []uint32{msg2.Seat, msg2.Biseat}
		others := t.GetOtherSeats(seats)

		t.broadcast4(seats, msg2)
		t.broadcast4(others, msg4)

		t.turnDetail(userid, "agree")

		t.DirtyCheck(seat, biseat)

		t.DeskAct.ActSeat = seat

		// 只有2人时增加剩余人机看牌概率
		t.addRobotSeeRateWhenJust2Player()
		t.pauseGame(PAUSE_REASON_1, 5, msg2.Loseseat)
		//延时3秒
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		// t.setNextActSeat()
	} else { //拒绝比牌
		msg1 := new(pb.JHCoinReplyBiRsp)
		t.send2userid(userid, msg1)

		next := t.getNextActSeat()

		msg := new(pb.JHCoinReplyBiNtf)
		msg.Seat = next
		msg.Biseat = t.getSeatid(userid)
		msg.Agree = false
		t.broadcast(msg)

		t.turnDetail(userid, "refuse")

		t.DeskAct.ActSeat = next

		// 记录比牌被拒绝
		t.getSeat(next).SideShowBeRejected = true

		t.setNextActSeat()

	}
}

// '比牌
func (t *Desk) coinBi(userid string) {
	t.chargeInGameCancel(userid)

	msg := new(pb.JHCoinBiRsp)

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

	status := t.getStatus(seatid)
	if status == nil {
		msg.Error = pb.Failed
		t.send2userid(userid, msg)
		return
	}
	// 记录玩家操作
	var act int32
	if t.DeskAct.ActState&int32(pb.ACT_SHOW) == int32(pb.ACT_SHOW) {
		act |= int32(pb.ACT_SHOW)
	} else if t.DeskAct.ActState&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) {
		act |= int32(pb.ACT_SIDESHOW)
	}
	if !t.isSee(seatid) {
		act |= int32(pb.ACT_BLIND)
	}
	status.ActHistory = append(status.ActHistory, act)

	num := t.ActAnte
	if t.isSee(seatid) {
		num = num * 2 //看牌double
	}

	// 私人房娱乐模式不判断
	if t.Rtype != int32(pb.ROOM_TYPE1) || t.Gmode != 1 {
		if user.GetScore() < num {
			msg.Error = pb.NotEnoughCoin
			msg.Amount, msg.GiveAmount = t.getRechargeAmount(seatid)
			msg.StrategyType = int32(t.strategyType)
			msg.ChargeInfo = t.getChargeInfo(seatid)
			t.send2userid(userid, msg)
			t.chargeInGameBegin(seatid)
			// t.coinFold(userid) //金币不足弃牌
			return
		}
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
		//TODO 扣除金币,比牌金额不足怎么处理
		if algo.HuaCompare(cs1, cs2) {
			//cs1 赢
			msg.Winseat = seatid
			msg.Loseseat = biseat
		} else {
			//cs2 赢
			msg.Winseat = biseat
			msg.Loseseat = seatid
		}

		msg3 := new(pb.JHCoinCallNtf)
		msg3.Info = t.setBet(seatid, userid, num, fmt.Sprintf("tp%s房间show", t.DeskData.Rid))
		t.recordBet(num)
		msg3.Pot = t.DeskGame.BetNum
		t.broadcast(msg3)

		if v, ok := t.DeskAct.ActSeats[msg.Loseseat]; ok {
			v.Alive = false
			t.Lose(msg.Loseseat)
		}
		//操作值中去掉操作
		// t.DeskHua.ActState ^= int32(pb.ACT_BI)
		t.send2userid(userid, msg)

		a1 := algo.HuaType(cs1)
		a2 := algo.HuaType(cs2)

		msg2 := new(pb.JHCoinBiNtf)
		msg2.State |= int32(pb.ACT_SHOW)
		msg2.Seat = msg.Seat
		msg2.Biseat = msg.Biseat
		msg2.Cards = cs1
		msg2.Bicards = cs2
		msg2.Type = a1
		msg2.Bitype = a2
		msg2.Winseat = msg.Winseat
		msg2.Loseseat = msg.Loseseat
		msg2.IsShow = true

		msg4 := new(pb.JHCoinBiNtf)
		utils.Clone(msg4, msg2)
		msg4.Cards = []uint32{}
		msg4.Bicards = []uint32{}

		seats := []uint32{msg2.Seat, msg2.Biseat}
		others := t.GetOtherSeats(seats)

		t.broadcast4(seats, msg2)
		t.broadcast4(others, msg4)

		t.turnDetail(userid, fmt.Sprintf("show:%d", num))

		t.DirtyCheck(seatid, biseat)

		t.pauseGame(PAUSE_REASON_1, 5, nil)
		// timerChan := time.After(3 * time.Second)
		// <-timerChan
		//结束或者继续
		// t.setNextActSeat()
	} else if t.DeskAct.ActState&int32(pb.ACT_SIDESHOW) == int32(pb.ACT_SIDESHOW) { //sideshow

		t.send2userid(userid, msg)

		seat := t.getSeatid(userid)

		// num := t.ActAnte * 2
		msg2 := new(pb.JHCoinCallNtf)
		msg2.Info = t.setBet(seat, userid, num, fmt.Sprintf("tp%s房间sideshow", t.DeskData.Rid))
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
	// userid := t.getRouter(ctx)
	// errcode := t.changeDeskCheck(userid)
	// if errcode != pb.OK {
	// 	rsp := new(pb.JHCoinChangeRoomRsp)
	// 	rsp.Error = errcode
	// 	ctx.Respond(rsp)
	// 	return
	// }
	// t.nnLeave(userid, ctx)
	// //匹配房间消息
	// msg := new(pb.ChangeDesk)
	// msg.Roomid = t.DeskData.Rid
	// msg.Rtype = t.DeskData.Rtype
	// msg.Gtype = t.DeskData.Gtype
	// msg.Ltype = t.DeskData.Ltype
	// msg.Dtype = t.DeskData.Dtype
	// msg.Userid = userid
	// msg.Sender = ctx.Sender() //玩家进程
	// nodePid.Request(msg, ctx.Self())
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

// 获取局内充值金额（旧）
func (t *Desk) getRechargeAmountDeprecated(seat uint32) (uint32, uint32) {
	if t.isStrategy && t.chargeAmount != 0 {
		return t.chargeAmount, uint32(float64(t.chargeAmount) * 0.05)
	} else {
		rechargeMap := t.DeskAct.ActRechargeTimes
		configs := t.Game.TP.RoomRecharge
		gives := t.Game.TP.RoomRechargeGive
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

// 获取玩家预计赢分
func (t *Desk) getEstimateWin(seat uint32) int64 {
	status := t.getStatus(seat)
	if status == nil {
		return 0
	}

	total := t.DeskGame.BetNum
	bi := t.ActAnte
	bet := status.ActNum

	return total + bi - bet
}

// 获取玩家预计可提现金
func (t *Desk) getEstimateCash(seat uint32) int64 {
	userid := t.getUserid(seat)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return 0
	}

	return t.getEstimateWin(seat) + role.OutDiamond
}

// 获取局内充值金额, 局内充值不赠送
func (t *Desk) getRechargeAmount(seat uint32) (uint32, uint32) {
	// 保存玩家充值金额及赠送金额
	// 剧情模式亏损库存在玩家充值中可能会变化, 导致回调查询赠送金额不一致, 这里第一次获取后缓存一下
	if amount, ok := t.chargingAmounts[seat]; ok && amount != nil {
		zlog.Infof("desk=%s user=%s read charge in game: %#v", t.Rid, t.getUserid(seat), amount)
		return amount.ChargeAmount, amount.GiveAmount
	}

	chargeAmount := new(data.TpChargeAmount)
	target := int32(t.getEstimateCash(seat))
	chargeAmount.Target = target

	conf := t.getRechargeAmountConfig(seat)
	zlog.Infof("局内充值: %s, %s, target=%d, %#v", t.GameId, t.getUserid(seat), target, conf.ValueRange)

	var chargeAmounts []int32 = conf.ValueRange  // 额度档位
	var viceAccountUp int32 = conf.ViceAccountUp // 小号升档

	if conf.StoryGameCharge && t.model == 3 {
		// 剧情局局内充值
		getRatio := func(wan int32) float64 {
			return float64(wan) / float64(10000)
		}

		if !conf.CanWithdraw {
			target = int32(float64(t.getEstimateWin(seat)) / getRatio(conf.ChargeInGameGiftRatio))
			zlog.Infof("剧情模式局内充值: %s, %s, %d -> %d, %#v", t.GameId, t.getUserid(seat), chargeAmount.Target, target, conf)
			chargeAmount.Target = target
		}
	}

	for i, v := range chargeAmounts {
		if v >= target {
			// 判断是否升档
			player := t.getPlayer(t.getUserid(seat))
			if player != nil && player.ViceAccount { // 小号强行升档
				chargeAmount.ViceAccountUp = viceAccountUp
				upNext := i + int(viceAccountUp)
				if upNext >= len(chargeAmounts) {
					upNext = len(chargeAmounts) - 1
				}

				zlog.Infof("小号强行升档: %s, %s, %d -> %d, %#v -> %#v, %v", t.GameId, t.getUserid(seat), i, upNext, v, chargeAmounts[upNext], chargeAmounts)
				v = chargeAmounts[upNext]
			}
			// return uint32(v), uint32(float64(v) * getRatio(conf.ChargeInGameGiftRatio))
			chargeAmount.ChargeAmount = uint32(v)
			t.chargingAmounts[seat] = chargeAmount
			glog.Debugf("desk=%s user=%s cache charge in game: %#v", t.GameId, t.getUserid(seat), chargeAmount)
			return chargeAmount.ChargeAmount, 0
		}
	}

	// 取最高档位
	chargeAmount.ChargeAmount = uint32(chargeAmounts[len(chargeAmounts)-1])
	t.chargingAmounts[seat] = chargeAmount
	glog.Debugf("desk=%s user=%s cache charge in game max: %#v", t.GameId, t.getUserid(seat), chargeAmount)
	return chargeAmount.ChargeAmount, 0
}

// 获取局内充值金额配置
func (t *Desk) getRechargeAmountConfig(seatid uint32) tb.TpChargeInGameValueRecord {
	//检查玩家模式
	checkModel := func(model int32, v *tb.TpChargeInGameValueRecord) bool {
		if v.Model == -1 {
			return true
		}
		if model == v.Model {
			return true
		}
		return false
	}

	//检查玩家阶段
	checkStage := func(stage int32, v *tb.TpChargeInGameValueRecord) bool {
		if v.Stage == -1 {
			return true
		}
		if stage == v.Stage {
			return true
		}
		return false
	}

	//检查今日充值次数区间
	checkChargeNum := func(num int32, v *tb.TpChargeInGameValueRecord) bool {
		if v.TodayChargeNumRange[0] == -1 && v.TodayChargeNumRange[1] == -1 {
			return true
		}
		if num >= v.TodayChargeNumRange[0] && num <= v.TodayChargeNumRange[1] {
			return true
		}
		return false
	}

	//检查今日充值金额区间
	checkChargeAmount := func(amount int64, v *tb.TpChargeInGameValueRecord) bool {
		if v.TodayChargeAmountRange[0] == -1 && v.TodayChargeAmountRange[1] == -1 {
			return true
		}
		if int32(amount) >= v.TodayChargeAmountRange[0] && int32(amount) <= v.TodayChargeAmountRange[1] {
			return true
		}
		return false
	}

	//检查今日提现次数区间
	checkWithdrawNum := func(num int32, v *tb.TpChargeInGameValueRecord) bool {
		if v.TodayWithdrawNumRange[0] == -1 && v.TodayWithdrawNumRange[1] == -1 {
			return true
		}
		if num >= v.TodayWithdrawNumRange[0] && num <= v.TodayWithdrawNumRange[1] {
			return true
		}
		return false
	}

	//检查今日提现金额区间
	checkWithdrawAmount := func(amount int64, v *tb.TpChargeInGameValueRecord) bool {
		if v.TodayWithdrawAmountRange[0] == -1 && v.TodayWithdrawAmountRange[1] == -1 {
			return true
		}
		if int32(amount) >= v.TodayWithdrawAmountRange[0] && int32(amount) <= v.TodayWithdrawAmountRange[1] {
			return true
		}
		return false
	}

	//检查盈利区间
	checkWinOrLose := func(amount int64, v *tb.TpChargeInGameValueRecord) bool {
		if v.WinOrLoseRange[0] == -1 && v.WinOrLoseRange[1] == -1 {
			return true
		}
		if int32(amount) >= v.WinOrLoseRange[0] && int32(amount) <= v.WinOrLoseRange[1] {
			return true
		}
		return false
	}

	//检查玩家系数区间
	checkPlayerFactor := func(factor float64, v *tb.TpChargeInGameValueRecord) bool {
		if v.PlayerFactorRange[0] == -1 && v.PlayerFactorRange[1] == -1 {
			return true
		}
		if factor >= v.PlayerFactorRange[0] && factor <= v.PlayerFactorRange[1] {
			return true
		}
		return false
	}

	//检查局内充值成功率
	checkChargeInGameSuccessRate := func(rate int32, v *tb.TpChargeInGameValueRecord) bool {
		if v.ChargeInGameSuccessRate == -1 {
			return true
		}
		if rate >= v.ChargeInGameSuccessRate {
			return true
		}
		return false
	}

	conf := table.GetTables().TpChargeInGameValueTable.GetDataList()

	userid := t.getUserid(seatid)
	role := t.getRole(userid)
	if userid == "" || role == nil {
		return *conf[0]
	}

	roomId, _ := strconv.ParseInt(t.Game.Id, 10, 32)
	factor := handler.GetFactor(role.User)
	rate := float64(role.TpUserChargeInGameNumOfSuccess) / float64(role.TpUserChargeInGameNumOfTrigger) * 10000

	for _, v := range conf {
		if v.Room == int32(roomId) &&
			checkModel(role.TpUserModel, v) &&
			checkStage(role.TpUserStage, v) &&
			checkChargeNum(role.TpUserTodayChargeNum, v) &&
			checkChargeAmount(role.TpUserTodayChargeAmount, v) &&
			checkWithdrawNum(role.TpUserTodayWithdrawNum, v) &&
			checkWithdrawAmount(role.TpUserTodayWithdrawAmount, v) &&
			checkWinOrLose(role.TpUserTotalWinOrLoseAmount, v) &&
			checkPlayerFactor(factor, v) &&
			checkChargeInGameSuccessRate(int32(rate), v) {
			return *v
		}
	}

	return *conf[0]
}

//.

// vim: set foldmethod=marker foldmarker=//',//.:
