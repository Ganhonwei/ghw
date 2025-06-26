package mines

import (
	"goserver/gen/pb"
	"goserver/pkg/game/event"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// nnLeave 玩家主动离开
func (t *Desk) nnLeave(ctx actor.Context) {
	msg := new(pb.MinesLeaveRsp)
	role := t.getRole()
	if role == nil {
		msg.Error = pb.NotInRoom
		glog.Errorf("mines leave not in room")
		return
	}
	msg.Userid = role.Userid

	errcode := t.leave()
	if errcode == pb.OK && t.autoRound > 0 {
		// 自动对局中
		errcode = pb.GameStartedCannotLeave
	}
	if errcode != pb.OK {
		msg.Error = errcode
		ctx.Respond(msg)
		return
	}
	ctx.Respond(msg)

	// 事件
	t.eventPost(event.Initiative_Leave, &event.InitiativeLeaveEvent{Gtype: int(pb.MINES)})
	t.eventPost(event.BreakingGift, new(event.BreakingGiftEvent)) // 破产礼包

	t.notifyGateUserLeft(pb.OK, int32(pb.OK))

	//清除数据
	t.offlineDesk(role.Userid)
	t.userLeaveDesk(role.Userid)
}

// leave 玩家离开
func (t *Desk) leave() pb.ErrCode {
	//离线处理
	if role := t.getRole(); role == nil {
		return pb.NotInRoom
	}

	//离开房间条件
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		if t.IsGaming() { //游戏进行中
			return pb.GameStartedCannotLeave
		} else { //游戏没开始
			return pb.OK
		}
	}
	return pb.OK
}

// 通知网关玩家离开
func (t *Desk) notifyGateUserLeft(err pb.ErrCode, reason int32) {
	if v := t.role; v != nil && v.Pid != nil {
		// if !v.Offline {
		msg := new(pb.LeftDesk)
		msg.Error = err
		msg.Gtype = int32(pb.MINES)
		msg.Reason = reason
		v.Pid.Tell(msg)
		// }
	}
}

// 踢除离线玩家
func (t *Desk) kickOffline() {
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0), //自由
		int32(pb.ROOM_TYPE1), //私人
		int32(pb.ROOM_TYPE2): //百人
		if t.role != nil {
			if !t.role.Offline {
				return
			}
			if errcode := t.leave(); errcode != pb.OK {
				return
			}
			//离开状态消息
			t.userLeaveDesk(t.role.Userid)
		}
	}
}

// 离开状态消息
func (t *Desk) userLeaveDesk(userid string) {
	if t.role == nil {
		return
	}

	//通知网关离开
	t.notifyGateUserLeft(pb.OK, int32(pb.OK))

	//清除数据
	t.role = nil

	msg := new(pb.LeaveDesk)
	msg.Userid = userid
	msg.Roomid = t.DeskData.Rid
	//离开房间
	nodePid.Tell(msg)

	//离开房间,TODO 玩家进程中操作,一致性
	t.roomPid.Request(msg, t.selfPid)

	// 关掉桌子
	t.selfPid.Tell(&pb.ServeStop{})
}
