package hua

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *Desk) CloseDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CloseDesk)
	glog.Debugf("CloseDesk %#v", arg)
	a.closeDesk(arg, ctx)
}

func (a *Desk) LeaveDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LeaveDesk)
	glog.Debugf("LeaveDesk %#v", arg)
	a.leaveDesk(arg, ctx)
}

func (a *Desk) SyncConfig(ctx actor.Context) {
	msg := ctx.Message()
	//更新配置
	arg := msg.(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	a.syncConfig(arg, ctx)
}

func (a *Desk) PrintDesk(ctx actor.Context) {
	//打印牌局状态信息,test
	a.printOver()
}

func (a *Desk) EnterDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.EnterDesk)
	// glog.Debugf("EnterDesk %#v", arg)
	a.enterDesk(arg, ctx)
}

func (a *Desk) OfflineDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.OfflineDesk)
	glog.Debugf("OfflineDesk %#v", arg)
	//离线消息
	a.offlineDesk(arg.Userid)
}

func (a *Desk) ChangeCurrency(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangeCurrency)
	//充值或购买同步
	a.changeCurrency(arg)
}

// 关闭桌子
func (a *Desk) closeDesk(arg *pb.CloseDesk, ctx actor.Context) {
	//TODO
	//响应
	//rsp := new(pb.ClosedDesk)
	//ctx.Respond(rsp)
}

//'离开房间

// 私人房房离开检查
func (t *Desk) privLeaveCheck(userid string) pb.ErrCode {
	if t.Rtype != int32(pb.ROOM_TYPE1) {
		return pb.OK
	}
	if t.IsGaming() {
		return pb.GameStartedCannotLeave
	}

	// 不在游戏中
	if !t.dismissOnRoundOver && // 不回合结束解散
		t.settleExitTime == 0 && // 不在解散倒计时中
		(len(t.seats) > 1 && t.Cid == userid) { // 房主不能退
		// 可以退
		return pb.GameStartedCannotLeave
	}
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

// 玩家掉线离开桌子(被动离开)
func (a *Desk) leaveDesk(arg *pb.LeaveDesk, ctx actor.Context) {
	//TODO 同一个房间?
	//if arg.Roomid != a.DeskData.Rid {
	//}
	//离线
	defer a.offlineDesk(arg.Userid)
	//响应消息
	rsp := new(pb.LeftDesk)
	if _, ok := a.roles[arg.Userid]; ok {
		errcode := a.leave(arg.Userid)
		// 私人房掉线策略 todo 等180秒
		if a.DeskData.Rtype == int32(pb.ROOM_TYPE1) {
			errcode = a.privLeaveCheck(arg.Userid)
		}
		rsp.Error = errcode
	}
	ctx.Respond(rsp)
	//成功离开后移除
	if rsp.Error == pb.OK {
		//TODO 私人房间掉线是否移除
		a.userLeaveDesk(arg.Userid)
	}
}

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
				// 不是重连
				user.SetFraction(int64(config.GetPvpRoom().TpFunInitScore))
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
	rsp.Gameid = a.Game.Id
	rsp.Roomid = a.DeskData.Rid
	rsp.Rtype = a.DeskData.Rtype
	rsp.Gtype = a.DeskData.Gtype
	rsp.Code = a.DeskData.Code

	rsp.Userid = user.GetUserid()
	rsp.Desk = ctx.Self()
	//ctx.Respond(rsp)
	arg.Sender.Tell(rsp)
	//TODO 优化
	if errcode == pb.AlreadyInRoom {
		return
	}
	// 机器人加入
	if user.Robot {
		a.robotCalling--
	} else {
		// 玩家加入
		a.isChangeTable = a.isChangeTable || arg.IsChangeTable

		if !arg.IsChangeTable {
			// 非换桌, 重置tp1003高潮涌现策略概率
			zlog.Errorf("%s reset 1003 hp: ", user.Userid)
			user.TpUserGcyxHp = 0
		}
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
		0, 0, 0, 0, 0, int32(pb.LOG_TYPE2), userid, fmt.Sprintf("TP加入%s房间", a.DeskData.Rid), "")
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
	switch a.Rtype {
	case int32(pb.ROOM_TYPE1): // 私人房
		switch a.chargePrevState {
		case int32(pb.STATE_BET):
			a.state = a.chargePrevState
			a.timer = 120
		case int32(pb.STATE_READY):
			a.privRoundChargingResume(false, userid)
		}
	default:
		a.state = int32(pb.STATE_BET)
		a.timer = 120
	}

	seatid := a.getSeatid(userid)

	rechargeMap := a.DeskAct.ActRechargeTimes
	rechargeMap[seatid] += 1

	msg := new(pb.ChargeInGameFinishNtf)
	msg.Seat = seatid
	msg.Totaltimer = int64(ChargeInGameTime-120) + utils.LocalTime().Unix()
	a.broadcast(msg)
	a.SetTpUserChargeInGameNumOfSuccess(seatid)

	if amount, ok := a.chargingAmounts[seatid]; ok {
		delete(a.chargingAmounts, seatid)
		zlog.Infof("局内充值成功: %s, %d, %#v", userid, a.model, amount)
	}

	// 局内充值实付次数
	if !a.isRobot(seatid) {
		a.isCharge = true

		for i := len(a.detail.ChargeInGame) - 1; i >= 0; i-- {
			if a.detail.ChargeInGame[i][0] == int32(seatid) {
				a.detail.ChargeInGame[i][3] = 1
			}
		}
	}
}

func (a *Desk) reqRoom(msg interface{}) interface{} {
	timeout := 3 * time.Second
	res1, err1 := a.roomPid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRoom err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

// 添加黑名单
func (a *Desk) AddBlackList(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AddBlackList)
	role := a.roles[arg.UserId]
	if role == nil {
		return
	}
	if role.User.Status == 1 {
		role.User.Status = 3
		return
	}
	// 如果是在黑名单就放出来
	role.User.Status = 1
}

func (a *Desk) GameRechargeAmount(ctx actor.Context) {
	// userid := a.getRouter(ctx)
	arg := ctx.Message().(*pb.GameRechargeAmount)
	rsp := new(pb.GameRechargeAmounted)
	seat := a.getSeatid(arg.Userid)
	rsp.Amount, rsp.GiveAmount = a.getRechargeAmount(seat)

	// 局内充值拉单记录
	for i := len(a.detail.ChargeInGame) - 1; i >= 0; i-- {
		if a.detail.ChargeInGame[i][0] == int32(seat) {
			// a.detail.ChargeInGame[i][1] = int32(rsp.Amount)
			a.detail.ChargeInGame[i][2] = 1
		}
	}
	ctx.Respond(rsp)
}

// vim: set foldmethod=marker foldmarker=//',//.:
