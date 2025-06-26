package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) JOKERCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinEnterRoomReq)
	glog.Debugf("JOKERCoinEnterRoomReq %#v", arg)
	// arg.Id = ""
	rs.enterJOKERCoin(arg, ctx)
}

func (rs *RoleActor) JOKERFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeEnterRoomReq)
	glog.Debugf("JOKERFreeEnterRoomReq %#v", arg)
	// rs.enterJOKERFree(arg, ctx)
}

func (rs *RoleActor) JOKERFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeDealerReq)
	glog.Debugf("JOKERFreeDealerReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERFreeDealerRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeDealerListReq)
	glog.Debugf("JOKERFreeDealerListReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERFreeDealerListRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERSitReq)
	glog.Debugf("JOKERSitReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERSitRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeBetReq)
	glog.Debugf("JOKERFreeBetReq %#v", arg)
	// rs.nnJOKERFreeBet(arg, ctx)
}

func (rs *RoleActor) JOKERFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeTrendReq)
	glog.Debugf("JOKERFreeTrendReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERFreeTrendRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeWinersReq)
	glog.Debugf("JOKERFreeWinersReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERFreeWinersRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERFreeRolesReq)
	glog.Debugf("JOKERFreeRolesReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERFreeRolesRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERRoomListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERRoomListReq)
	glog.Debugf("JOKERRoomListReq %#v", arg)
	rs.getJOKERRoomList(arg, ctx)
}

func (rs *RoleActor) JOKEREnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKEREnterRoomReq)
	glog.Debugf("JOKEREnterRoomReq %#v", arg)
	// rs.enterJOKERPriv(arg, ctx)
}

func (rs *RoleActor) JOKERCreateRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCreateRoomReq)
	glog.Debugf("JOKERCreateRoomReq %#v", arg)
	// rs.createJOKERRoom(arg, ctx)
}

func (rs *RoleActor) JOKERLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERLeaveReq)
	glog.Debugf("JOKERLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERReadyReq(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.JOKERReadyReq)
	// glog.Debugf("JOKERReadyReq %#v", arg)
	// if rs.gamePid == nil {
	// 	rsp := new(pb.JOKERReadyRsp)
	// 	rsp.Error = pb.NotInRoom
	// 	rs.Send(rsp)
	// 	return
	// }
	// rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERReady2Req)
	glog.Debugf("JOKERReady2Req %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERReady2Rsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// func (rs *RoleActor) ChatTextReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.ChatTextReq)
// 	glog.Debugf("ChatTextReq %#v", arg)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.ChatTextRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

func (rs *RoleActor) JOKERGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERGameRecordReq)
	glog.Debugf("JOKERGameRecordReq %#v", arg)
	//TODO
}

func (rs *RoleActor) JOKERLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERLaunchVoteReq)
	glog.Debugf("JOKERLaunchVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERLaunchVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERVoteReq)
	glog.Debugf("JOKERVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinSeeReq)
	glog.Debugf("JOKERCoinSeeReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinSeeRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinCallReq)
	glog.Debugf("JOKERCoinCallReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinCallRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinRaiseReq)
	glog.Debugf("JOKERCoinRaiseReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinRaiseRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinFoldReq)
	glog.Debugf("JOKERCoinFoldReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinFoldRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinBiReq)
	glog.Debugf("JOKERCoinBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinReplyBiReq)
	glog.Debugf("JOKERCoinReplyBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinReplyBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JOKERCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JOKERCoinChangeRoomReq)
	glog.Debugf("JOKERCoinChangeRoomReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JOKERCoinChangeRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入百人房间
// func (rs *RoleActor) enterJOKERFree(arg *pb.JOKERFreeEnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterJOKERMatchDesk(ctx)
// 	if msg != nil {
// 		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入私人房间
// func (rs *RoleActor) enterJOKERPriv(arg *pb.JOKEREnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterJOKERMatchDesk(ctx)
// 	if msg != nil {
// 		//msg.Rtype = int32(pb.ROOM_TYPE1) //私人
// 		msg.Rtype = int32(pb.ROOM_TYPE0) //自由
// 		msg.Code = arg.Code              //邀请码
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入自由房间
func (rs *RoleActor) enterJOKERCoin(arg *pb.JOKERCoinEnterRoomReq, ctx actor.Context) {
	if rs.gamePid != nil {
		game1 := config.GetGame(arg.Gameid)
		game2 := config.GetGame(rs.gameId)
		if game1.Gtype != game2.Gtype { //想进的游戏和已经在的游戏不同
			rsp := new(pb.JOKERCoinEnterRoomRsp)
			rsp.Error = pb.InOtherGame
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			rsp.GameType = game2.Gtype
			rsp.RoomType = int32(game2.RoomType)
			rs.Send(rsp)
			return
		}
	}

	msg := rs.enterJOKERMatchDesk(ctx)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE0) //自由
		msg.Gameid = arg.Gameid          //游戏ID
		msg.Roomid = arg.Roomid          //房间ID
		// msg.Dtype = int32(pb.DESK_TYPE1) //玩法类型
		//计算出匹配房间等级,算法一致
		//msg.Ltype = int32(pb.ROOM_LEVEL1) //等级
		msg.Ltype = handler.MatchLevel(rs.User.GetCoin())
		if msg.Ltype < 0 {
			rsp := new(pb.JOKERCoinEnterRoomRsp)
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
		if !rs.Robot {
			//判断最低进入
			game := config.GetGame(arg.Gameid)
			if rs.User.GetScore() < int64(game.Min_Access) {
				rsp := new(pb.JOKERCoinEnterRoomRsp)
				rsp.Error = pb.NotEnoughCoin
				rs.Send(rsp)
				return
			}
			//判断最高进入
			if rs.User.GetScore() > int64(game.Max_Access) && game.Max_Access != -1 {
				rsp := new(pb.JOKERCoinEnterRoomRsp)
				rsp.Error = pb.TooManyCoin
				rs.Send(rsp)
				return
			}
		}

		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterJOKERMatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.joker").Name()
	msg.Gtype = int32(pb.JOKER) //金花
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getJOKERRoomList(arg *pb.JOKERRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.joker").Name()
	msg.Gtype = int32(pb.HUA) //金花
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
// func (rs *RoleActor) nnJOKERFreeBet(arg *pb.JOKERFreeBetReq, ctx actor.Context) {
// 	if rs.User.IsTourist() {
// 		rsp := new(pb.JOKERFreeBetRsp)
// 		rsp.Error = pb.TouristInoperable
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.gamePid == nil {
// 		rsp := new(pb.JOKERFreeBetRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	value := arg.GetValue()
// 	seat := arg.GetSeat()
// 	if !(seat >= uint32(pb.DESK_SEAT2) &&
// 		seat <= uint32(pb.DESK_SEAT9)) {
// 		rsp := new(pb.JOKERFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if value <= 0 {
// 		rsp := new(pb.JOKERFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.User.GetCoin() < int64(value) {
// 		rsp := new(pb.JOKERFreeBetRsp)
// 		rsp.Error = pb.NotEnoughCoin
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 创建房间
// func (rs *RoleActor) createJOKERRoom(arg *pb.JOKERCreateRoomReq, ctx actor.Context) {
// 	//TODO 验证
// 	msg := &pb.CreateDesk{
// 		Rname:    arg.Rname,
// 		Dtype:    arg.Dtype,
// 		Ante:     arg.Ante,
// 		Round:    arg.Round,
// 		Payment:  arg.Payment,
// 		Count:    arg.Count,
// 		Pub:      arg.Pub,
// 		Minimum:  int64(arg.Minimum),
// 		Maximum:  int64(arg.Maximum),
// 		Mode:     arg.Mode,
// 		Multiple: arg.Multiple,
// 		//TODO 消耗
// 		Cost: 100,
// 	}
// 	switch msg.Dtype {
// 	case int32(pb.DESK_TYPE0):
// 	case int32(pb.DESK_TYPE1):
// 	case int32(pb.DESK_TYPE2):
// 	default:
// 		msg.Dtype = int32(pb.DESK_TYPE0)
// 	}
// 	msg.Name = cfg.Section("game.hua").Name()
// 	msg.Gtype = int32(pb.HUA)        //金花
// 	msg.Rtype = int32(pb.ROOM_TYPE1) //私人
// 	msg.Cid = rs.User.GetUserid()
// 	msg.Sender = ctx.Self()
// 	//节点中匹配
// 	rs.dbmsPid.Request(msg, ctx.Self())
// }
