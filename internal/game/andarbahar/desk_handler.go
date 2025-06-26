package andarbahar

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

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

// 私人房房离开检查 掉线、离线超时
func (t *Desk) privLeaveCheck(userid string) pb.ErrCode {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return pb.OK
	}
	// 在游戏中
	if t.IsGaming() {
		return pb.GameStartedCannotLeave
	}

	if !t.dismissOnRoundOver && // 不回合结束解散
		t.settleExitTime == 0 && // 不在解散倒计时中
		len(t.seats) > 1 && t.Cid == userid { // 还有人房主掉线不能退
		return pb.GameStartedCannotLeave
	}
	// 可以退
	return pb.LeaveEarly
}

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
		// 私人房掉线策略 todo 等180秒
		if errcode == pb.OK && a.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
			errcode = a.privLeaveCheck(arg.Userid)
		}
		rsp.Error = errcode
	}
	ctx.Respond(rsp)
	//成功离开后移除
	if rsp.Error == pb.OK {
		//TODO 私人房间掉线是否移除
		a.userLeaveDesk(arg.Userid)
		// 通知自己回到大厅
		ntf := &pb.UPLeaveNtf{
			Userid: arg.Userid,
		}
		a.send2userid(arg.Userid, ntf)
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

	//对战房
	if a.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
		_, reconnect := a.roles[user.GetUserid()]
		if !reconnect {
			// 娱乐模式初始分数
			if a.DeskData.Gmode == 1 {
				user.SetFraction(int64(config.GetPvpRoom().AbFunInitScore))
			} else {
				// 真金初始携带金额
				if _, ok := a.roles[user.GetUserid()]; !ok {
					if !user.Robot && user.GetScore() < int64(a.DeskData.Game.MinFirstEntry) {
						rsp.Error = pb.NotEnoughDiamond
						glog.Errorf("entry Desk err: %d", rsp.Error)
						arg.Sender.Tell(rsp)
						return
					}
				}
				// 真金模式最低准入
				if !user.Robot && user.GetScore() < int64(a.DeskData.Game.Min_Access) {
					rsp.Error = pb.NotEnoughDiamond
					glog.Errorf("entry Desk err: %d", rsp.Error)
					arg.Sender.Tell(rsp)
					return
				}
			}
		}
	} else {
		//点控和新手一个人一张桌子
		if arg.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
			r, _ := a.roleCountNum()
			if r >= 1 {
				// 有人了，不能进了
				rsp.Error = pb.RoomNotExist
				ctx.Respond(rsp)
				return
			}
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

		// 对战房房主为庄家
		if a.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
			if len(a.seats) == 1 {
				for k, seat := range a.seats {
					a.DeskGame.Dealer = seat.Userid
					a.DeskGame.DealerSeat = k
				}
			}
		}
	}
	//响应消息
	rsp.Roomid = a.DeskData.Rid
	rsp.Gameid = a.DeskData.Game.Id
	rsp.Rtype = a.DeskData.Rtype
	rsp.Gtype = a.DeskData.Gtype
	rsp.Code = a.DeskData.Code

	rsp.Userid = user.GetUserid()
	rsp.Desk = ctx.Self()
	arg.Sender.Tell(rsp)
	if arg.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
		ctx.Respond(rsp)
	}
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
		0, 0, 0, 0, 0, int32(pb.LOG_TYPE2), userid, fmt.Sprintf("andarbahar进入%s房间", a.DeskData.Rid), "")
	a.send2userid(userid, msg2)
}

//.

// '同步配置
func (a *Desk) syncConfig(arg *pb.SyncConfig, ctx actor.Context) {
	b := make(map[string]data.Game)
	err = json.Unmarshal(arg.Data, &b)
	if err != nil {
		glog.Errorf("syncConfig Unmarshal err %v", err)
		return
	}
	for _, v := range b {
		if a.DeskData.Unique == v.Id {
			//TODO 只更新可变内容
			//a.DeskData.Ante = v.Ante
			//a.DeskData.Chip = v.Chip
			deskData := handler.NewDeskData(&v)
			deskData.Rid = a.DeskData.Rid
			a.DeskData = deskData
			return
		}
	}
}

//.

// 更新货币
func (a *Desk) changeCurrency(arg *pb.ChangeCurrency) {
	user := a.getPlayer(arg.Userid)
	if user == nil {
		glog.Debugf("changeCurrency err %s", arg.Userid)
		return
	}
	user.AddCurrency(arg.Diamond, arg.Coin, arg.Card, arg.Chip, arg.Give, arg.Out)
	if arg.Money > 0 {
		user.AddMoney(uint32(arg.Money))
		a.chargeInGameFinish(arg.Userid)
	}
}

// 局内充值完成
func (a *Desk) chargeInGameFinish(userid string) {
	if a.state != int32(pb.STATE_CHARGE) {
		return
	}
	seatid := a.getSeat(userid)

	switch a.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		a.privRoundChargingResume(false, userid)
	}

	rechargeMap := a.ActRechargeTimes
	rechargeMap[seatid] += 1

	msg := new(pb.ChargeInGameFinishNtf)
	msg.Seat = seatid
	msg.Totaltimer = int64(ChargeInGameTime-120) + utils.LocalTime().Unix()
	a.broadcast(msg)

	a.isCharge = true
}

func (a *Desk) CloseServer(ctx actor.Context) {
	a.closeServer = true
}

// vim: set foldmethod=marker foldmarker=//',//.:
