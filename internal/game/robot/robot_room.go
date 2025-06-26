package robot

import (
	"errors"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 召唤机器人
func (a *RoleActor) RobotMsg(ctx actor.Context) {
	arg := ctx.Message().(*pb.RobotMsg)
	glog.Debugf("RobotMsg %v", arg)
	a.buildUserInfo(arg)
	a.enterRoom(arg, ctx)
	// 表情
	var emoji_config tb.EmojiRobotEmojiConfigRecord
	err := json.Unmarshal(arg.Emoji, &emoji_config)
	if err == nil {
		a.emoji = emoji_config
	}
}

func (a *RoleActor) EnteredDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.EnteredDesk)
	glog.Debugf("EnteredDesk %v", arg)
	err := a.enteredDesk(arg, ctx)
	if err != nil {
		a.closeRs()
	}
}

func (a *RoleActor) MatchedDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.MatchedDesk)
	glog.Debugf("MatchedDesk %v", arg)
	err := a.matchedDesk(arg, ctx)
	if err != nil {
		a.closeRs()
	}
}

func (a *RoleActor) PushCurrencyNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.PushCurrencyNtf)
	glog.Debugf("PushCurrencyNtf %v", arg)
	a.Coin += arg.Data.Coin
}

// 局内充值通知
func (a *RoleActor) ChargeInGameNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChargeInGameNtf)
	if a.seat == arg.Seat {
		a.chargeInGame = true
		glog.Infof("robot charge in game: %d, %s", a.seat, a.Userid)
	}
}

func (a *RoleActor) ChargeInGameFinishNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChargeInGameFinishNtf)
	if arg.Seat == a.seat {
		// 重放机器人假装局内充值前动作
		switch a.gtype {
		case int32(pb.HUA):
			if a.prevRAction != nil {
				a.huaRunRAction(a.prevRAction)
				// a.prevRAction = nil
			} else {
				a.huaRunAction(a.prevAction)
			}
		case int32(pb.HUA2):
			a.huaRunAction(a.prevAction)
		case int32(pb.JOKER):
			a.jokerRunAction(a.prevAction)
		case int32(pb.AK47):
			a.ak47RunAction(a.prevAction)
		default:
			glog.Error("unknown robot charge finish in game type: gtype=%v, action=%v", a.gtype, a.prevAction)
		}
	}
}

// 进入房间
func (rs *RoleActor) enterRoom(arg *pb.RobotMsg, ctx actor.Context) {
	msg := new(pb.MatchDesk)
	switch arg.Gtype {
	case int32(pb.HUA):
		msg = rs.enterJHMatchDesk(arg)
	case int32(pb.LHD):
		msg = rs.matchlhdDesk(arg)
	case int32(pb.SEVEN):
		msg = rs.matchUPDesk(arg)
	case int32(pb.RUMMY):
		msg = rs.enterRummyMatchDesk(arg)
	case int32(pb.AK47):
		msg = rs.enterAK47MatchDesk(arg)
	case int32(pb.JOKER):
		msg = rs.enterJokerMatchDesk(arg)
	case int32(pb.CRASH):
		msg = rs.matchCRASHDesk(arg)
	case int32(pb.ABAR):
		msg = rs.matchabDesk(arg)
	case int32(pb.LOTTERY):
		msg = rs.matchcpDesk(arg)
	case int32(pb.PLANE):
		msg = rs.matchPLANEDesk(arg)
	case int32(pb.RUMMY2):
		msg = rs.enterRummy2MatchDesk(arg)
	case int32(pb.HUA2):
		msg = rs.enterJH2MatchDesk(arg)
	}
	rs.selectDesk(msg, ctx)
}

// 查找不同类型房间远程节点
func (rs *RoleActor) selectDesk(msg *pb.MatchDesk, ctx actor.Context) {
	msg.Sender = ctx.Self()
	switch msg.Gtype {
	case int32(pb.HUA), int32(pb.HUA2), int32(pb.JOKER), int32(pb.AK47), int32(pb.RUMMY), int32(pb.RUMMY2):
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
			rs.dbmsPid.Request(msg, ctx.Self())
		default:
			glog.Errorf("selectDesk match fail %#v", msg)
		}
	case int32(pb.LHD), int32(pb.SEVEN), int32(pb.CRASH),
		int32(pb.ABAR), int32(pb.LOTTERY), int32(pb.PLANE):
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
func (rs *RoleActor) matchedDesk(arg *pb.MatchedDesk, ctx actor.Context) error {
	if arg.Error != pb.OK {
		//失败消息
		glog.Errorf("matched desk filed arg %#v", arg)
		return errors.New("match desk fail")
	}
	if arg.Desk == nil {
		//失败消息
		glog.Errorf("matched desk filed arg %#v", arg)
		return errors.New("match desk fail")
	}
	msg := new(pb.EnterDesk)
	msg.Gameid = arg.Gameid
	msg.Roomid = arg.Roomid
	msg.Gtype = arg.Gtype
	msg.Rtype = arg.Rtype
	msg.Dtype = arg.Dtype
	msg.Ltype = arg.Ltype
	// if rs.SimRobot {
	// 	msg.New = true
	// }
	if !rs.enterDeskMsg(msg, ctx) {
		//失败消息
		glog.Errorf("matched desk filed arg %#v", arg)
		return errors.New("match desk fail")
	}
	//请求消息
	arg.Desk.Request(msg, ctx.Self())
	return nil
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

// 加入游戏结果
func (rs *RoleActor) enteredDesk(arg *pb.EnteredDesk, ctx actor.Context) error {
	if arg.Error != pb.OK {
		//失败消息
		glog.Errorf("entered desk filed arg %#v", arg)
		return errors.New("enter desk fail")
	}
	if arg.Desk == nil {
		//失败消息
		arg.Error = pb.EnterFail
		glog.Errorf("entered desk filed arg %#v", arg)
		return errors.New("enter desk fail")
	}
	rs.gamePid = arg.Desk
	rs.gameId = arg.Gameid
	rs.roomId = arg.Roomid
	rs.gtype = arg.Gtype
	//加入成功后获取房间数据
	if rs.enterdDeskMsg(arg, ctx) {
		return nil
	}
	return errors.New("enter desk fail")
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
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.LHD):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LHFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.SEVEN):
		switch msg.Rtype {
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
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.ABFreeEnterRoomReq) //加入消息
			msg2.Roomid = msg.Roomid
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.LOTTERY):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.LotteryEnterReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	case int32(pb.PLANE):
		switch msg.Rtype {
		case int32(pb.ROOM_TYPE2): //百人
			msg2 := new(pb.PLANEEnterRoomReq) //加入消息
			rs.gamePid.Request(msg2, ctx.Self())
		default:
			glog.Errorf("enterdDesk match fail %#v", msg)
		}
	default:
		glog.Errorf("enterdDesk match fail %#v", msg)
	}
	return true
}
