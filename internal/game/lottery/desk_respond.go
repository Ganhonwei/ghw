package lottery

import (
	"strings"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// external function

// ' 打印
func (t *Desk) printOver() {
	glog.Debugf("game over data -> %#v", t.DeskData)
	glog.Debugf("game over roles -> %d", len(t.roles))
	glog.Debugf("game over router -> %d", len(t.router))
	glog.Debugf("game over state -> %d", t.state)
	// glog.Debugf("game over dealer -> %s", t.DeskGame.Dealer)
	// glog.Debugf("game over dealer seat -> %d", t.DeskGame.DealerSeat)
	for k, v := range t.roles {
		if v.GetRobot() {
			continue
		}
		glog.Debugf("game over userid %s , %v", k, v.Seat, v.Offline)
	}
	// for k, v := range t.seats {
	// 	glog.Debugf("game over seat %d -> %s", k, v.Userid)
	// }

	// glog.Debugf("game over Carry -> %d", t.DeskFree.Carry)
	// glog.Debugf("game over DealerNum -> %d", t.DeskFree.DealerNum)
	glog.Debugf("game over Cards -> %#v", t.LCards)
	glog.Debugf("game over Bets -> %#v", t.DeskFree.Bets)
	glog.Debugf("game over SeatBets -> %#v", t.DeskFree.SeatBets)
	// glog.Debugf("game over Multiple -> %#v", t.DeskFree.Multiple)
	// glog.Debugf("game over Score1 -> %#v", t.DeskFree.Score1)
	// glog.Debugf("game over Score2 -> %#v", t.DeskFree.Score2)
	// glog.Debugf("game over Score3 -> %#v", t.DeskFree.Score3)
	if t.DeskPriv != nil {
		glog.Debugf("game over priv vote %d", t.DeskPriv.VoteSeat)
		glog.Debugf("game over priv score %#v", t.DeskPriv.PrivScore)
		glog.Debugf("game over priv joins %#v", t.DeskPriv.Joins)
	}
}

//.

// ' 聊天
func (t *Desk) chatText(arg *pb.ChatTextReq, ctx actor.Context) {
	userid := t.getRouter(ctx)
	seat := t.getSeat(userid)
	glog.Debugf("ChatTextReq %s, %d", userid, seat)
	//房间消息广播,聊天
	msg := handler.ChatTextMsg(seat, userid, arg.Content)
	msg.Error = t.chatEmoji(userid, arg.GetContent())
	if msg.Error != pb.OK {
		t.send2userid(userid, msg)
		return
	}
	t.broadcast(msg)
}

// func (t *Desk) chatVoiceReq(arg *pb.ChatVoiceReq, ctx actor.Context) {
// 	userid := t.getRouter(ctx)
// 	seat := t.getSeat(userid)
// 	glog.Debugf("CChatVoiceReq %s, %d", userid, seat)
// 	//房间消息广播,聊天
// 	// t.broadcast(handler.ChatVoiceReqMsg(seat, userid, arg.Content))
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

//.

// ' enter 进入桌子
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
		return pb.AlreadyInRoom
	}
	//加入游戏
	p := new(data.DeskRole)
	p.User = user
	p.Pid = pid
	// p.CashPro = handler.GetCashProportion(user) // 彩金占比
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
	var i uint32
	for i = 1; i <= t.DeskData.Count; i++ {
		if _, ok := t.seats[i]; !ok {
			p.Seat = i
			t.seats[i] = &data.DeskSeat{
				Userid: p.User.GetUserid(),
			}
			break
		}
	}
}

// TODO 不同类型房间进入限制
func (t *Desk) enterCheck(user *data.User) pb.ErrCode {
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
	if user.GetCoin() < t.DeskData.Maximum {
		return pb.NotEnoughCoin
	}
	return pb.OK
}

//.

//' 离开房间处理

// nnLeave 玩家主动离开,TODO 下注也可以离开
func (t *Desk) nnLeave(userid string, ctx actor.Context) {
	//离线
	defer t.offlineDesk(userid)
	errcode := t.leave(userid, 0)
	glog.Debugf("cpLeave userid %s, err %v", userid, errcode)
	rsp := new(pb.LotteryLeaveRsp)
	t.send2userid(userid, rsp)
	t.eventPost(userid, event.BreakingGift, new(event.BreakingGiftEvent)) // 破产礼包
	if errcode == pb.OK {
		t.userLeaveDesk(userid)
		return
	}
	t.userLeaveDeskMsg(userid, 0)
}

// leave 玩家离开,TODO 下注也可以离开
func (t *Desk) leave(userid string, err int32) pb.ErrCode {
	//离线处理 TODO
	// defer t.userLeaveDeskMsg(userid)
	if _, ok := t.roles[userid]; !ok {
		return pb.NotInRoom
	}
	if _, ok := t.DeskFree.Bets[userid]; ok && t.state != int32(pb.STATE_OVER) { //下注中
		return pb.GameStartedCannotLeave
	}
	t.userLeaveDeskMsg(userid, err)
	return pb.OK
}

// 离开状态消息
func (t *Desk) userLeaveDeskMsg(userid string, reason int32) {
	if v, ok := t.roles[userid]; ok {
		if !v.Offline {
			msg2 := new(pb.LeftDesk)
			msg2.Gtype = int32(pb.LOTTERY)
			msg2.Reason = reason
			v.Pid.Tell(msg2)
		}
	}
}

// 离开状态消息
func (t *Desk) userLeaveDesk(userid string) {
	if v, ok := t.roles[userid]; ok {
		// 事件
		t.eventPost(userid, event.Initiative_Leave, &event.InitiativeLeaveEvent{Gtype: int(pb.LOTTERY)})
		//清除数据
		delete(t.seats, v.Seat)
		delete(t.roles, userid)
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

// vim: set foldmethod=marker foldmarker=//',//.:
