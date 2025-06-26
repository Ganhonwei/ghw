package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) AK47CoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinEnterRoomReq)
	glog.Debugf("AK47CoinEnterRoomReq %#v", arg)
	// arg.Id = ""
	rs.enterAK47Coin(arg, ctx)
}

func (rs *RoleActor) AK47FreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeEnterRoomReq)
	glog.Debugf("AK47FreeEnterRoomReq %#v", arg)
	// rs.enterAK47Free(arg, ctx)
}

func (rs *RoleActor) AK47FreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeDealerReq)
	glog.Debugf("AK47FreeDealerReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47FreeDealerRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47FreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeDealerListReq)
	glog.Debugf("AK47FreeDealerListReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47FreeDealerListRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47SitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47SitReq)
	glog.Debugf("AK47SitReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47SitRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47FreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeBetReq)
	glog.Debugf("AK47FreeBetReq %#v", arg)
	// rs.nnAK47FreeBet(arg, ctx)
}

func (rs *RoleActor) AK47FreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeTrendReq)
	glog.Debugf("AK47FreeTrendReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47FreeTrendRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47FreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeWinersReq)
	glog.Debugf("AK47FreeWinersReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47FreeWinersRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47FreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47FreeRolesReq)
	glog.Debugf("AK47FreeRolesReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47FreeRolesRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47RoomListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47RoomListReq)
	glog.Debugf("AK47RoomListReq %#v", arg)
	rs.getAK47RoomList(arg, ctx)
}

func (rs *RoleActor) AK47EnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47EnterRoomReq)
	glog.Debugf("AK47EnterRoomReq %#v", arg)
	// rs.enterAK47Priv(arg, ctx)
}

func (rs *RoleActor) AK47CreateRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CreateRoomReq)
	glog.Debugf("AK47CreateRoomReq %#v", arg)
	// rs.createAK47Room(arg, ctx)
}

func (rs *RoleActor) AK47LeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47LeaveReq)
	glog.Debugf("AK47LeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47LeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47ReadyReq(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.AK47ReadyReq)
	// glog.Debugf("AK47ReadyReq %#v", arg)
	// if rs.gamePid == nil {
	// 	rsp := new(pb.AK47ReadyRsp)
	// 	rsp.Error = pb.NotInRoom
	// 	rs.Send(rsp)
	// 	return
	// }
	// rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47Ready2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47Ready2Req)
	glog.Debugf("AK47Ready2Req %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47Ready2Rsp)
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

func (rs *RoleActor) AK47GameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47GameRecordReq)
	glog.Debugf("AK47GameRecordReq %#v", arg)
	//TODO
}

func (rs *RoleActor) AK47LaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47LaunchVoteReq)
	glog.Debugf("AK47LaunchVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47LaunchVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47VoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47VoteReq)
	glog.Debugf("AK47VoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47VoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinSeeReq)
	glog.Debugf("AK47CoinSeeReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinSeeRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinCallReq)
	glog.Debugf("AK47CoinCallReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinCallRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinRaiseReq)
	glog.Debugf("AK47CoinRaiseReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinRaiseRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinFoldReq)
	glog.Debugf("AK47CoinFoldReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinFoldRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinBiReq)
	glog.Debugf("AK47CoinBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinReplyBiReq)
	glog.Debugf("AK47CoinReplyBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinReplyBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) AK47CoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AK47CoinChangeRoomReq)
	glog.Debugf("AK47CoinChangeRoomReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.AK47CoinChangeRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入百人房间
// func (rs *RoleActor) enterAK47Free(arg *pb.AK47FreeEnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterAK47MatchDesk(ctx)
// 	if msg != nil {
// 		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入私人房间
// func (rs *RoleActor) enterAK47Priv(arg *pb.AK47EnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterAK47MatchDesk(ctx)
// 	if msg != nil {
// 		//msg.Rtype = int32(pb.ROOM_TYPE1) //私人
// 		msg.Rtype = int32(pb.ROOM_TYPE0) //自由
// 		msg.Code = arg.Code              //邀请码
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入自由房间
func (rs *RoleActor) enterAK47Coin(arg *pb.AK47CoinEnterRoomReq, ctx actor.Context) {
	if rs.gamePid != nil {
		game1 := config.GetGame(arg.Gameid)
		game2 := config.GetGame(rs.gameId)
		if game1.Gtype != game2.Gtype { //想进的游戏和已经在的游戏不同
			rsp := new(pb.AK47CoinEnterRoomRsp)
			rsp.Error = pb.InOtherGame
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			rsp.GameType = game2.Gtype
			rsp.RoomType = int32(game2.RoomType)
			rs.Send(rsp)
			return
		}
	}

	msg := rs.enterAK47MatchDesk(ctx)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE0) //自由
		msg.Gameid = arg.Gameid          //游戏ID
		msg.Roomid = arg.Roomid          //房间ID
		// msg.Dtype = int32(pb.DESK_TYPE1) //玩法类型
		//计算出匹配房间等级,算法一致
		//msg.Ltype = int32(pb.ROOM_LEVEL1) //等级
		msg.Ltype = handler.MatchLevel(rs.User.GetCoin())
		if msg.Ltype < 0 {
			rsp := new(pb.AK47CoinEnterRoomRsp)
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
		if !rs.Robot {
			//判断最低进入
			game := config.GetGame(arg.Gameid)
			if rs.User.GetScore() < int64(game.Min_Access) {
				rsp := new(pb.AK47CoinEnterRoomRsp)
				rsp.Error = pb.NotEnoughCoin
				rs.Send(rsp)
				return
			}
			//判断最高进入
			if rs.User.GetScore() > int64(game.Max_Access) && game.Max_Access != -1 {
				rsp := new(pb.AK47CoinEnterRoomRsp)
				rsp.Error = pb.TooManyCoin
				rs.Send(rsp)
				return
			}
		}

		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterAK47MatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.ak47").Name()
	msg.Gtype = int32(pb.AK47) //金花
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getAK47RoomList(arg *pb.AK47RoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.ak47").Name()
	msg.Gtype = int32(pb.AK47) //金花
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
// func (rs *RoleActor) nnAK47FreeBet(arg *pb.AK47FreeBetReq, ctx actor.Context) {
// 	if rs.User.IsTourist() {
// 		rsp := new(pb.AK47FreeBetRsp)
// 		rsp.Error = pb.TouristInoperable
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.gamePid == nil {
// 		rsp := new(pb.AK47FreeBetRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	value := arg.GetValue()
// 	seat := arg.GetSeat()
// 	if !(seat >= uint32(pb.DESK_SEAT2) &&
// 		seat <= uint32(pb.DESK_SEAT9)) {
// 		rsp := new(pb.AK47FreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if value <= 0 {
// 		rsp := new(pb.AK47FreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.User.GetCoin() < int64(value) {
// 		rsp := new(pb.AK47FreeBetRsp)
// 		rsp.Error = pb.NotEnoughCoin
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 创建房间
// func (rs *RoleActor) createAK47Room(arg *pb.AK47CreateRoomReq, ctx actor.Context) {
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
