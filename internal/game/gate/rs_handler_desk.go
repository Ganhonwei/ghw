package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/event"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) EnteredDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.EnteredDesk)
	glog.Debugf("EnteredDesk %#v", arg)
	rs.enterdDesk(arg, ctx)
}

func (rs *RoleActor) MatchedDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.MatchedDesk)
	glog.Debugf("MatchedDesk %#v", arg)
	rs.matchedDesk(arg, ctx)
}

func (rs *RoleActor) CreatedDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CreatedDesk)
	glog.Debugf("CreatedDesk %#v", arg)
	rs.createdDesk(arg, ctx)
}

func (rs *RoleActor) LeftDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LeftDesk)
	glog.Debugf("LeftDesk %#v, userid %s", arg, rs.User.GetUserid())
	if arg.Gtype == int32(pb.LOTTERY) {
		rs.cpPid = nil
	} else if arg.Error == pb.OK || arg.Error == pb.LeaveEarly {
		rs.gamePid = nil
		rs.gameId = ""
		rs.gtype = 0
		rs.roomId = ""
		rs.roomCode = ""
	}

	// 离开游戏后到大厅上报
	if rs.gamePid == nil && rs.cpPid == nil {
		if err = mq.NatsPublish(mq.TopicGameToLobby, &pb.PublishGameToLobby{
			UserPid: rs.pid,
			Userid:  rs.Userid,
			Gtype:   arg.Gtype,
			Ts:      time.Now().Unix(),
		}); err != nil {
			glog.Error("publish user left desk to lobby error", err)
		}
	}

	rs.ET_NewbieGuide2(arg)
	// 礼包弹窗
	// rs.checkGiftPop()
	rs.pbGift()
	// rs.breakingGift()

	if arg.Reason == int32(pb.Kick105) {
		rs.Kick105Flag = true
		rs.status = true
	} else if arg.Reason == int32(pb.Kick200) {
		rs.Kick200Flag = true

		rs.status = true
	} else if arg.Reason == int32(pb.KickCloseServer) {
		rs.pid.Tell(new(pb.CloseServer))
	}
	// 私人房解散且用户在离线状态,添加标志待用户重连时提示
	if arg.Reason == int32(pb.PrivRoomDismissed) &&
		!rs.User.OnlineStatus {
		rs.privRoomDismissed = true
	}
}

func (rs *RoleActor) SetRecord(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.SetRecord)
	glog.Debugf("SetRecord %#v", arg)
	if rs.User != nil {
		rs.status = true
		rs.User.SetRecord(arg.Rtype)
	}
	// if rs.User.GetAgent() != "" && (rs.User.Win+rs.User.Lost+rs.User.Ping) == 10 {
	// 	//更新有效代理绑定
	// 	msg := handler.AgentBuildUpdateMsg(rs.User.GetAgent(), rs.User.GetUserid(), 0, 1, 0)
	// 	rs.rolePid.Tell(msg)
	// }
}

func (rs *RoleActor) GotRoomList(ctx actor.Context) {
	// arg := msg.(*pb.GotRoomList)
	// glog.Debugf("GotRoomList %#v", arg)
	// msg2 := new(pb.SNNRoomList)
	// rs.Send(msg2)
}

func (rs *RoleActor) ChangedDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangedDesk)
	glog.Debugf("ChangedDesk %#v", arg)
	rs.changedDesk(arg, ctx)
}

// 加入游戏结果
func (rs *RoleActor) enterdDesk(arg *pb.EnteredDesk, ctx actor.Context) {
	if arg.Error != pb.OK {
		//失败消息
		rs.enterdDeskErr(arg, ctx)
		glog.Errorf("entered desk filed arg %#v", arg)
		return
	}
	if arg.Desk == nil {
		//失败消息
		arg.Error = pb.EnterFail
		rs.enterdDeskErr(arg, ctx)
		glog.Errorf("entered desk filed arg %#v", arg)
		return
	}
	if arg.Gtype == int32(pb.LOTTERY) {
		msg2 := new(pb.LotteryEnterReq) //加入消息
		msg2.RoomId = arg.Roomid
		arg.Desk.Request(msg2, ctx.Self())
		rs.cpPid = arg.Desk
		return
	}
	rs.gamePid = arg.Desk
	rs.gameId = arg.Gameid
	rs.gtype = arg.Gtype
	rs.roomId = arg.Roomid
	rs.roomCode = arg.Code
	//加入成功后获取房间数据
	if rs.enterdDeskMsg(arg, ctx) {
		return
	}
}

// 进入房间数据,返回房间数据
func (rs *RoleActor) enterdDeskMsg(msg *pb.EnteredDesk, ctx actor.Context) bool {
	if rs.gamePid == nil {
		return false
	}
	//区分类型消息
	switch msg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JHCoinEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.JHEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		// case int32(pb.ROOM_TYPE2): //百人
		// 	msg2 := new(pb.JHFreeEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.JOKER):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JOKERCoinEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.AK47):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.AK47CoinEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.RMCoinEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.RMEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		// case int32(pb.ROOM_TYPE2): //百人
		// 	msg2 := new(pb.JHFreeEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.LHD):
		switch msg.Rtype {
		// case int32(pb.ROOM_TYPE0): //自由
		// 	msg2 := new(pb.JHCoinEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		// case int32(pb.ROOM_TYPE1): //私人
		// 	msg2 := new(pb.JHEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LHFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.SEVEN):
		switch msg.Rtype {
		// case int32(pb.ROOM_TYPE0): //自由
		// 	msg2 := new(pb.JHCoinEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		// case int32(pb.ROOM_TYPE1): //私人
		// 	msg2 := new(pb.JHEnterRoomReq) //加入消息
		// 	rs.gamePid.Request(msg2, ctx.Self())
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.UPFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.CRASH):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.CRASHEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.ABAR):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.ABEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.ABFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.PLANE):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.PLANEEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.REDBLACK):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.RBFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.MINES):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.MinesEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGems2EnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGemsEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	default:
		glog.Errorf("enterdDesk match fail %#v", msg)
	}
	return true
}

// 查找不同类型房间远程节点
func (rs *RoleActor) selectDesk(msg *pb.MatchDesk, ctx actor.Context) {
	msg.Sender = ctx.Self()
	switch msg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2), int32(pb.JOKER), int32(pb.AK47), int32(pb.RUMMY), int32(pb.RUMMY2), int32(pb.MINES),
		int32(pb.FORTUNE_GEMS2), int32(pb.FORTUNE_GEMS):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			if msg.Gameid != "" && msg.Roomid != "" { //存在游戏ID和房间ID直接查找
				rs.roomPid.Request(msg, ctx.Self())
			} else if msg.Gameid != "" { //或者不存在时去节点中匹配
				rs.dbmsPid.Request(msg, ctx.Self())
			}
		case int32(pb.ROOM_TYPE1): //私人
			rs.roomPid.Request(msg, ctx.Self())
		case int32(pb.ROOM_TYPE2): //百人
			glog.Errorf("selectDesk match fail2 %#v", msg)
			rs.dbmsPid.Request(msg, ctx.Self())
		default:
			glog.Errorf("selectDesk match fail %#v", msg)
		}
	case int32(pb.LHD), int32(pb.SEVEN), int32(pb.CRASH), int32(pb.ABAR),
		int32(pb.LOTTERY), int32(pb.PLANE), int32(pb.REDBLACK):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
			rs.roomPid.Request(msg, ctx.Self())
		case int32(pb.ROOM_TYPE2): //百人
			rs.dbmsPid.Request(msg, ctx.Self())
		default:
			glog.Errorf("selectDesk match fail %#v", msg)
		}
	default:
		glog.Errorf("selectDesk match fail %#v", msg)
	}
}

// 配置远程节点结果,然后加入游戏
func (rs *RoleActor) matchedDesk(arg *pb.MatchedDesk, ctx actor.Context) {
	if arg.Error != pb.OK {
		//失败消息
		rs.matchedDeskErr(arg, ctx)
		glog.Errorf("matched desk filed arg %#v", arg)
		return
	}
	if arg.Desk == nil {
		//失败消息
		arg.Error = pb.MatchFail
		rs.matchedDeskErr(arg, ctx)
		glog.Errorf("matched desk filed arg %#v", arg)
		return
	}
	msg := new(pb.EnterDesk)
	msg.Gameid = arg.Gameid
	msg.Roomid = arg.Roomid
	msg.Gtype = arg.Gtype
	msg.Rtype = arg.Rtype
	msg.Dtype = arg.Dtype
	msg.Ltype = arg.Ltype
	msg.Userid = rs.Userid
	msg.IsChangeTable = arg.IsChangeTable
	// if rs.SimRobot {
	// 	msg.New = true
	// }
	if !rs.enterDeskMsg(msg, ctx) {
		//失败消息
		rs.matchedDeskErr(arg, ctx)
		glog.Errorf("matched desk filed arg %#v", arg)
		return
	}
	if msg.Gtype == int32(pb.LOTTERY) {
		// 彩票节点
		rs.cpPid = arg.Desk
	}

	//请求消息
	arg.Desk.Request(msg, ctx.Self())
}

// 已经在游戏中,直接加入
func (rs *RoleActor) enterGame(ctx actor.Context) bool {
	if rs.gamePid != nil {
		msg := new(pb.EnterDesk)
		if !rs.enterDeskMsg(msg, ctx) {
			glog.Errorf("userid %s enter faild %s",
				rs.User.GetUserid(), rs.gamePid.String())
		}
		rs.gamePid.Request(msg, ctx.Self())
		return true
	}
	return false
}

// 加入房间消息
func (rs *RoleActor) enterDeskMsg(msg *pb.EnterDesk, ctx actor.Context) bool {
	result4, err4 := json.Marshal(rs.User)
	if err4 != nil {
		glog.Errorf("user Marshal err %v", err4)
		return false
	}
	//玩家数据
	msg.Sender = ctx.Self()
	msg.Data = result4
	return true
}

// 响应加入游戏失败消息
func (rs *RoleActor) enterdDeskErr(msg *pb.EnteredDesk, ctx actor.Context) {
	switch msg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JHCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.JHEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.JHFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.AK47):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.AK47CoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.AK47EnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.AK47FreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.JOKER):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JOKERCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.JOKEREnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.JOKERFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.LHD):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LHFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.SEVEN):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.UPFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.CRASH):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.CRASHEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.PLANE):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.PLANEEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.RMCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.RMEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.RMFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.LOTTERY):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LotteryEnterRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.ABAR):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.ABFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.ABEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.REDBLACK):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.RBFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.MINES):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.MinesEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGems2EnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGemsEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	default:
		glog.Errorf("enter Desk match fail %#v", msg)
	}
	if rs.gamePid != nil {
		rs.gamePid = nil
		rs.gameId = ""
		rs.gtype = 0
		rs.roomId = ""
		rs.roomCode = ""
	}
}

// 响应匹配游戏失败消息
func (rs *RoleActor) matchedDeskErr(msg *pb.MatchedDesk, ctx actor.Context) {
	switch msg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JHCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.JHEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.JHFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.AK47):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.AK47CoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.AK47EnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.AK47FreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.JOKER):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.JOKERCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.JOKEREnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.JOKERFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.LHD):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LHFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.SEVEN):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.UPFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.CRASH):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.CRASHEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.PLANE):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.PLANEEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.REDBLACK):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.RBFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.RMCoinEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE1): //私人
			msg2 := new(pb.RMEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.RMFreeEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.MINES):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.MinesEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS2):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGems2EnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	case int32(pb.FORTUNE_GEMS):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE0): //自由
			msg2 := new(pb.FortuneGemsEnterRoomRsp) //加入消息
			msg2.Error = msg.Error
			rs.Send(msg2)
		default:
			glog.Errorf("enter Desk match fail %#v", msg)
		}
	default:
		glog.Errorf("enter Desk match fail %#v", msg)
	}
}

// 创建房间结果
func (rs *RoleActor) createdDesk(arg *pb.CreatedDesk, ctx actor.Context) {
	if arg.Error != pb.OK {
		glog.Errorf("created desk filed arg %#v", arg)
		rs.createdDeskMsg(arg.Gtype, arg.Error)
		return
	}
	msg := new(pb.EnterDesk)
	msg.Gtype = arg.Gtype
	msg.Rtype = arg.Rtype
	if !rs.enterDeskMsg(msg, ctx) || arg.Desk == nil {
		rs.createdDeskMsg(arg.Gtype, arg.Error)
		return
	}
	arg.Desk.Request(msg, ctx.Self())
}

// 失败消息
func (rs *RoleActor) createdDeskMsg(Gtype int32, errcode pb.ErrCode) {
	switch Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		rsp := new(pb.JHCreateRoomRsp)
		rsp.Error = pb.CreateRoomFail
		rs.Send(rsp)
	case int32(pb.RUMMY), int32(pb.RUMMY2):
		rsp := new(pb.RMCreateRoomRsp)
		rsp.Error = pb.CreateRoomFail
		rs.Send(rsp)
	case int32(pb.ABAR):
		rsp := new(pb.ABCreateRoomRsp)
		rsp.Error = pb.CreateRoomFail
		rs.Send(rsp)
	default:
		glog.Errorf("matched DeskErr fail %d, %d", Gtype, errcode)
	}
}

// 换房间结果
func (rs *RoleActor) changedDesk(arg *pb.ChangedDesk, ctx actor.Context) {
	if arg.Error != pb.OK || arg.Desk == nil {
		glog.Errorf("changed failed %#v, userid %s", arg, rs.User.GetUserid())
		//失败消息
		rs.changedDeskMsg(arg, pb.ChangeFailed)
		return
	}
	msg := new(pb.EnterDesk)
	if !rs.enterDeskMsg(msg, ctx) {
		glog.Errorf("changed failed %#v, userid %s", arg, rs.User.GetUserid())
		//失败消息
		rs.changedDeskMsg(arg, pb.ChangeFailed)
		return
	}
	//请求消息
	arg.Desk.Request(msg, ctx.Self())
	rs.changedDeskMsg(arg, pb.OK)
}

func (rs *RoleActor) changedDeskMsg(arg *pb.ChangedDesk, errcode pb.ErrCode) {
	switch arg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2):
		rsp := new(pb.JHCoinChangeRoomRsp)
		rsp.Error = errcode
		rs.Send(rsp)
	}
}

// 总游戏时长
func (rs *RoleActor) LogGameTime(ctx actor.Context) {
	arg := ctx.Message().(*pb.LogGameTime)
	rs.TotalGameTime += uint64(arg.Time)

	beginner := table.GetTables().NewbieTable.Get()
	if beginner == nil {
		return
	}

	if rs.TotalGameTime >= uint64(beginner.Transfer[0])*60 && rs.Money == 0 { //总游戏时长平民条件
		rs.State = 3
		rs.status = true

		if rs.gamePid != nil {
			rs.gamePid.Request(&pb.UserGameState{State: 3, Userid: rs.Userid}, ctx.Self())
		}
		if rs.cpPid != nil {
			rs.cpPid.Request(&pb.UserGameState{State: 3, Userid: rs.Userid}, ctx.Self())
		}
		return
	}

	// if rs.RegistArea == 1 && rs.Withdrawable200 {
	// 	// B类,可提现到达200后，再玩10分钟变平民
	// 	rs.SurplusGameTime -= arg.Time
	// 	if rs.SurplusGameTime <= 0 {
	// 		rs.State = 3
	// 		rs.status = true
	// 	}
	// }
}

// 输赢记录
func (rs *RoleActor) FreeSetRecord(ctx actor.Context) {
	arg := ctx.Message().(*pb.FreeSetRecord)
	glog.Debugf("FreeSetRecord %#v", arg)
	user := rs.User
	wins := make([]data.FreeWin, 0)
	if user.FreeWinMap == nil {
		user.FreeWinMap = make(map[int32][]data.FreeWin)
	}
	if w, ok := user.FreeWinMap[arg.Gtype]; ok {
		wins = w
	}
	wins = append(wins, data.FreeWin{Wtype: arg.Rtype, Score: arg.Score})
	size := len(wins)
	if size > 20 {
		wins = wins[size-20:]
	}

	user.Round++
	user.FreeWinMap[int32(arg.Gtype)] = wins
	user.LastGType = arg.Gtype
	if arg.GameTime > 0 {
		user.OnlineTime += uint64(arg.GameTime)
		user.TotalOnlineTime += uint64(arg.GameTime)
	}

	beginner := table.GetTables().NewbieTable.Get()
	if beginner != nil && user.Round >= uint32(beginner.Transfer[1]) && rs.Money == 0 {
		// 新手转平民
		rs.State = 3
		if rs.gamePid != nil {
			rs.gamePid.Request(&pb.UserGameState{State: 3, Userid: rs.Userid}, ctx.Self())
		}
		if rs.cpPid != nil {
			rs.cpPid.Request(&pb.UserGameState{State: 3, Userid: rs.Userid}, ctx.Self())
		}
	}

	if rs.NewBiewConvertNormal() {
		// 通知
		msg := &pb.UserGameState{
			Userid: user.Userid,
			State:  int32(user.State),
		}
		if rs.gamePid != nil {
			rs.gamePid.Request(msg, ctx.Self())
		}
		if rs.cpPid != nil {
			rs.cpPid.Request(msg, ctx.Self())
		}
	}

	if rs.RoundGames == nil {
		rs.RoundGames = make(map[int32]int32)
	}
	rs.RoundGames[arg.Gtype]++

	rs.ET_FirstAndSecondRound(arg)
	rs.ET_NewbieGuide1(arg)

	// 打码量
	rs.vbFlow(arg.Bet)

	// 限时礼包事件
	if handler.IsFreeGame(rs.gtype) {
		rs.breakingGift()
	}
	// ntf := event.Event(user, event.LIMITED_GIFT, new(event.LimitedGiftEvent))
	// if ntf != nil {
	// 	rs.Send(ntf)
	// }

	// 玩游戏抽奖
	ntf := event.Event(user, event.PLAY_SHARE, &event.PlayShareEvent{Ptype: 1})
	if ntf != nil {
		rs.Send(ntf)
	}

	// 投注记录
	if rs.RegistArea == 1 {
		rs.BetRecord = append(rs.BetRecord, arg.Bet)
		if len(rs.BetRecord) > 200 {
			rs.BetRecord = rs.BetRecord[len(rs.BetRecord)-200:]
		}
		// 计算起付线
		rs.FluctuateLine = handler.FluctuateLine(rs.User)
		// if rs.Withdrawable200 && rs.State == data.NoveiceState {
		// 	rs.SurplusGame--
		// 	if rs.SurplusGame <= 0 {
		// 		rs.State = 3
		// 	}
		// }
	}

	// 输赢回合
	if rs.WinRates == nil {
		rs.WinRates = make(map[int32]data.WinLoseRound)
	}
	if b, ok := rs.WinRates[arg.Gtype]; ok {

		if arg.Score > 0 {
			b.WinRound++
		} else {
			b.LoseRound++
		}
		rs.WinRates[arg.Gtype] = b
	} else {
		b = data.WinLoseRound{}
		if arg.Score > 0 {
			b.WinRound++
		} else {
			b.LoseRound++
		}
		rs.WinRates[arg.Gtype] = b
	}

	rs.lhdStrategy(arg)
	rs.sevenUpStrategy(arg)

	rs.status = true

}

// tp类累计局数
func (rs *RoleActor) TPTotalRound(ctx actor.Context) {
	arg := ctx.Message().(*pb.TPTotalRound)
	glog.Debugf("TPTotalRound %#v", arg)
	if arg.Isreset {
		rs.ResetTPTotalRound()
	} else {
		rs.IncreTPTotalRound()
	}
	rs.status = true
}

// tp类触发次数
func (rs *RoleActor) TPTriggeTimes(ctx actor.Context) {
	arg := ctx.Message().(*pb.TPTriggeTimes)
	glog.Debugf("TPTriggeTimes %#v", arg)
	if arg.Isreset {
		rs.ResetTPTiggerTimes()
	} else {
		rs.IncreTPTiggerTimes()
	}
	rs.status = true
}

func (rs *RoleActor) TriggerStategy100(ctx actor.Context) {
	arg := ctx.Message().(*pb.TriggerStategy100)
	glog.Debugf("TriggerStategy100 %#v", arg)
	rs.SetStrategy100()
	rs.status = true
}

func (rs *RoleActor) TriggerStategy200(ctx actor.Context) {
	arg := ctx.Message().(*pb.TriggerStategy200)
	glog.Debugf("TriggerStategy200 %#v", arg)
	rs.SetStrategy200()
	rs.status = true
}

func (rs *RoleActor) TriggerFreeWelfare(ctx actor.Context) {
	arg := ctx.Message().(*pb.TriggerFreeWelfare)
	glog.Debugf("TriggerFreeWelfare %#v", arg)
	rs.FreeWelfare++
	if rs.cpPid != nil {
		rs.cpPid.Request(arg, ctx.Self())
	}
	if rs.gamePid != nil {
		rs.gamePid.Request(arg, ctx.Self())
	}
	rs.status = true
}

func (rs *RoleActor) getFreeDtype() int32 {
	user := rs.User
	if user.PCSwitch {
		return int32(pb.DESK_TYPE_POINTCONTROL)
	}
	switch user.State {
	case data.NoveiceState, data.ExceptionState, data.FrothState: // 新手,平民
		return int32(pb.DESK_TYPE_NEWBIEW)
	}
	if user.RegistArea == 1 {
		return int32(pb.DESK_TYPE_NORMAL_B)
	}
	return int32(pb.DESK_TYPE_NORMAL)
}

func (rs *RoleActor) PrivEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PrivEnterRoomReq)
	glog.Debugf("PrivEnterRoomReq %#v", arg)
	// 已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return
	}

	rsp := new(pb.PrivEnterRoomRsp)
	// 被踢1分钟内不能再加入
	if utils.Timestamp()-60 < rs.User.BeKickoutTime {
		rsp.Error = pb.EnterBeKickoutTime
		rs.Send(rsp)
		return
	}

	// 判断进入房间基本充值金额
	pvpRoom := config.GetPvpRoom()
	if rs.User.GetMoney() < uint32(pvpRoom.PvpChargeLimit) { // && rs.RegistArea == 0 {
		rsp.Error = pb.PvpRechargeRequire
		rsp.RechargeRequire = uint32(pvpRoom.PvpChargeLimit)
		rs.Send(rsp)
		return
	}

	rs.roomPid.Request(arg, ctx.Self())
}

// 玩家被踢出
func (rs *RoleActor) KickoutedRoom(ctx actor.Context) {
	arg := ctx.Message().(*pb.KickoutedRoom)
	glog.Debugf("KickoutedRoom %#v", arg)
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE1):
		// 更新被踢出时间
		rs.User.BeKickoutTime = utils.Timestamp()
	}
}

// 私人房牌局结束再来一局发起投票
func (rs *RoleActor) PrivLaunchAgainReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PrivLaunchAgainReq)
	glog.Debugf("PrivLaunchAgainReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.PrivLaunchAgainRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 私人房牌局结束再来一局投票
func (rs *RoleActor) PrivAgainReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PrivAgainReq)
	glog.Debugf("PrivAgainReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.PrivAgainRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 阶段变化同步
func (rs *RoleActor) TpUserStageChangeSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserStageChangeSync)
	if rs.TpUserStageChange(arg.Stage) {
		rs.status = true
	}
}

// 局数同步
func (rs *RoleActor) TpUserRoundSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserRoundSync)
	_ = arg
	rs.TpUserRound++
	rs.status = true
}

// 模式同步
func (rs *RoleActor) TpUserModelSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserModelSync)
	rs.TpUserModel = arg.Model
	rs.status = true
}

// 输赢记录点同步
func (rs *RoleActor) TpWinOrLoseAmountSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpWinOrLoseAmountSync)
	rs.TpUserWinOrLoseAmount = arg.Amount
	rs.status = true
}

// 变化修正值同步
func (rs *RoleActor) TpUserChangeCorrectionValueSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserChangeCorrectionValueSync)
	rs.TpUserChangeCorrectionValue = arg.Value
	rs.status = true
}

// 累计输赢额同步
func (rs *RoleActor) TpUserTotalWinOrLoseAmountSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserTotalWinOrLoseAmountSync)
	rs.TpUserTotalWinOrLoseAmount += arg.Amount
	rs.status = true
}

// 局内充值触发数
func (rs *RoleActor) TpUserChargeInGameNumOfTriggerSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserChargeInGameNumOfTriggerSync)
	_ = arg
	rs.TpUserChargeInGameNumOfTrigger++
	rs.status = true
}

// 局内充值成功数
func (rs *RoleActor) TpUserChargeInGameNumOfSuccessSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserChargeInGameNumOfSuccessSync)
	_ = arg
	rs.TpUserChargeInGameNumOfSuccess++
	rs.status = true
}

// 剧情局局内充值成功数
func (rs *RoleActor) TpUserChargeInGameStorySync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserChargeInGameStorySync)
	_ = arg
	rs.TpUserChargeInGameStory++
	rs.status = true
}

// 剧情局局内充值成功数
func (rs *RoleActor) TpUserStoryPlusSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserStoryPlusSync)
	_ = arg
	rs.TpUserStoryPlus = true
	rs.status = true
}

// 跟牌率触发数
func (rs *RoleActor) TpUserFollowRateTriggerSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserFollowRateTriggerSync)

	if rs.TpUserFollowRateTrigger == nil {
		rs.TpUserFollowRateTrigger = make(map[int32]int32)
	}

	if _, ok := rs.TpUserFollowRateTrigger[arg.CardType]; !ok {
		rs.TpUserFollowRateTrigger[arg.CardType] = 1
	} else {
		rs.TpUserFollowRateTrigger[arg.CardType]++
	}
	rs.status = true
}

// 跟牌率成功数
func (rs *RoleActor) TpUserFollowRateSuccessSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserFollowRateSuccessSync)

	if rs.TpUserFollowRateSuccess == nil {
		rs.TpUserFollowRateSuccess = make(map[int32]int32)
	}

	if _, ok := rs.TpUserFollowRateSuccess[arg.CardType]; !ok {
		rs.TpUserFollowRateSuccess[arg.CardType] = 1
	} else {
		rs.TpUserFollowRateSuccess[arg.CardType]++
	}
	rs.status = true
}

// 剧情cd
func (rs *RoleActor) TpUserStoryCDSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserStoryCDSync)

	if rs.TpUserStoryCD == nil {
		rs.TpUserStoryCD = make(map[int32]int32)
	}
	rs.TpUserStoryCD[arg.Id] = arg.Cd
	rs.status = true
}

// 剧情cd减少
func (rs *RoleActor) TpUserStoryCDReduce(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserStoryCDReduce)
	_ = arg
	if rs.TpUserStoryCD == nil {
		rs.TpUserStoryCD = make(map[int32]int32)
	}

	for k, v := range rs.TpUserStoryCD {
		if v > 0 {
			rs.TpUserStoryCD[k]--
		}
	}
	rs.status = true
}

// 控制策略历史
func (rs *RoleActor) TpUserControlStrategyHistorySync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserControlStrategyHistorySync)
	rs.TpUserControlStrategyHistory = append(rs.TpUserControlStrategyHistory, arg.Id)
	rs.status = true
}

// 游戏时长
func (rs *RoleActor) TpUserGameTimeSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserGameTimeSync)
	if rs.LastChargeTime != 0 {
		rs.TpUserGameTime += arg.Time
		rs.status = true
	}
}

// 今日游戏局数
func (rs *RoleActor) TpUserTodayGameRoundSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserTodayGameRoundSync)
	_ = arg
	rs.TpUserTodayGameRound++
	rs.status = true
}

// 控制策略今日触发数
func (rs *RoleActor) TpUserTodayControlStrategyNumSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.TpUserTodayControlStrategyNumSync)
	if rs.TpUserTodayControlStrategyNum == nil {
		rs.TpUserTodayControlStrategyNum = make(map[int32]int32)
	}
	if _, ok := rs.TpUserTodayControlStrategyNum[arg.Id]; !ok {
		rs.TpUserTodayControlStrategyNum[arg.Id] = 1
	} else {
		rs.TpUserTodayControlStrategyNum[arg.Id]++
	}

	rs.status = true
}

// 新手提现标记
func (rs *RoleActor) GetWithdrawFlag() {
	if rs.State != data.NoveiceState && rs.State != data.ExceptionState {
		return
	}

	conf := table.GetTables().TransferDemoCashTable.Get()
	level := len(conf.NewbieCarry)
	for i, v := range conf.NewbieCarry {
		if rs.OutDiamond < int64(v) {
			level = i
			break
		}
	}
	rs.KickWithdrawFlag = level
	rs.WithdrawPOPTime = utils.BsonNow().Unix()
}
