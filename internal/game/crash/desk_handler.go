package crash

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 关闭桌子
func (a *Desk) closeDesk(arg *pb.CloseDesk, ctx actor.Context) {
	//TODO
	//响应
	//rsp := new(pb.ClosedDesk)
	//ctx.Respond(rsp)
}

//'离开房间

// 掉线
func (a *Desk) offlineDesk(userid string) {
	a.setOffline(userid, true)
	pid := a.getPid(userid)
	if pid != nil {
		delete(a.router, pid.String())
	}
	//离线消息
	a.offlineMsg(userid)
}

// 玩家掉线离开桌子
func (a *Desk) leaveDesk(arg *pb.LeaveDesk, ctx actor.Context) {
	//TODO 同一个房间?
	//if arg.Roomid != a.DeskData.Rid {
	//}
	//离线
	defer a.offlineDesk(arg.Userid)
	//响应消息
	rsp := new(pb.LeftDesk)
	if _, ok := a.roles[arg.Userid]; ok {
		errcode := a.leave(arg.Userid, 0)
		rsp.Error = errcode
	}
	ctx.Respond(rsp)
	//成功离开后移除
	if rsp.Error == pb.OK {
		//TODO 私人房间掉线是否移除
		a.userLeaveDesk(arg.Userid)
	}
}

//.

// '进入房间
func (a *Desk) enterDesk(arg *pb.EnterDesk, ctx actor.Context) {
	rsp := new(pb.EnteredDesk)
	rsp.Gtype = arg.Gtype
	rsp.Rtype = arg.Rtype
	user := new(data.User)
	err2 := json.Unmarshal(arg.Data, user)
	if err2 != nil {
		glog.Errorf("user Unmarshal err %v", err2)
		rsp.Error = pb.RoomNotExist
		//ctx.Respond(rsp)
		arg.Sender.Tell(rsp)
		return
	}
	//点控桌子
	if arg.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
		r, _ := a.roleCountNum()
		if r >= 1 {
			// 有人了，不能进了
			rsp.Error = pb.RoomNotExist
			ctx.Respond(rsp)
			return
		}
	}
	//加入桌子
	errcode := a.enter(user, arg.Sender)
	if errcode != pb.OK && errcode != pb.AlreadyInRoom {
		glog.Errorf("entry Desk err: %d", errcode)
		rsp.Error = errcode
		//ctx.Respond(rsp)
		arg.Sender.Tell(rsp)
		return
	}
	if errcode == pb.OK {
		//加入房间消耗
		a.enterDeskCost(user.GetUserid())
	}
	//响应消息
	rsp.Roomid = a.DeskData.Rid
	rsp.Gameid = pb.GameType_name[int32(pb.CRASH)] // a.DeskData.Game.Id
	rsp.Rtype = a.DeskData.Rtype
	rsp.Gtype = a.DeskData.Gtype
	rsp.Userid = user.GetUserid()
	rsp.Desk = ctx.Self()
	if arg.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
		ctx.Respond(rsp)
	}
	arg.Sender.Tell(rsp)
	//TODO 优化
	if errcode == pb.AlreadyInRoom {
		return
	}
	//进入消息
	msg3 := new(pb.JoinDesk)
	msg3.Roomid = a.DeskData.Rid
	msg3.Rtype = a.DeskData.Rtype
	msg3.Gtype = a.DeskData.Gtype
	msg3.Userid = user.Userid
	msg3.Sender = arg.Sender
	nodePid.Request(msg3, ctx.Self())
	a.roomPid.Request(msg3, ctx.Self())
}

// AA消耗房间
func (a *Desk) isAADesk() bool {
	switch a.DeskData.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		return a.DeskData.Payment == 1 //AA
	}
	return false
}

// 加入房间消耗
func (a *Desk) enterDeskCost(userid string) {
	if !a.isAADesk() || userid == a.DeskData.Cid {
		return
	}
	msg2 := handler.ChangeCurrencyMsg(int64(a.DeskData.Cost),
		0, 0, 0, 0, 4, int32(pb.LOG_TYPE2), userid, fmt.Sprintf("CRASH进入%s房间", a.DeskData.Rid), "")
	a.send2userid(userid, msg2)
}

//.

// '同步配置
func (a *Desk) syncConfig(arg *pb.SyncConfig, ctx actor.Context) {
	// b := make(map[string]data.Game)
	// err = json.Unmarshal(arg.Data, &b)
	// if err != nil {
	// 	glog.Errorf("syncConfig Unmarshal err %v", err)
	// 	return
	// }

	// for _, v := range b {
	// 	if a.DeskData.Unique == v.Id {
	// 		//TODO 只更新可变内容
	// 		//a.DeskData.Ante = v.Ante
	// 		//a.DeskData.Chip = v.Chip
	// 		deskData := handler.NewDeskData(&v)
	// 		deskData.Rid = a.DeskData.Rid
	// 		a.DeskData = deskData
	// 		return
	// 	}
	// }
}

//.

// 更新货币
func (a *Desk) changeCurrency(arg *pb.ChangeCurrency) {
	user := a.getPlayer(arg.Userid)
	if user == nil {
		glog.Debugf("changeCurrency err %s", arg.Userid)
		return
	}
	if arg.Money > 0 {
		user.AddMoney(uint32(arg.Money))
	}
	user.AddCurrency(arg.Diamond, arg.Coin, arg.Card, arg.Chip, arg.Give, arg.Out)
}

func (a *Desk) CloseServer(ctx actor.Context) {
	a.closeServer = true
}

// vim: set foldmethod=marker foldmarker=//',//.:
