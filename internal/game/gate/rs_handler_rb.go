package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) RBFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBFreeEnterRoomReq)
	glog.Debugf("RBFreeEnterRoomReq %#v", arg)
	rs.enterRBFree(arg, ctx)
}

// 下注
func (rs *RoleActor) RBFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBFreeBetReq)
	// glog.Debugf("RBFreeBetReq %#v", arg)
	rs.rbFreeBet(arg, ctx)
}

// 离开
func (rs *RoleActor) RBLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.RBLeaveReq)
	glog.Debugf("RBLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RBLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入百人房间
func (rs *RoleActor) enterRBFree(arg *pb.RBFreeEnterRoomReq, ctx actor.Context) {
	msg := rs.enterRbMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterRbMatchDesk(ctx actor.Context, rid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	// if rs.gamePid != nil {
	// 	game2 := config.GetGame(rs.gameId)
	// 	if game2.Gtype != int32(pb.REDBLACK) {
	// 		rsp := new(pb.RBFreeEnterRoomRsp)
	// 		rsp.Error = pb.InOtherGame
	// 		rsp.Gameid = rs.gameId
	// 		rsp.Roomid = rs.roomId
	// 		rsp.GameType = game2.Gtype
	// 		rsp.RoomType = int32(game2.RoomType)
	// 		rs.Send(rsp)
	// 		return nil
	// 	}
	// }
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rb").Name()
	msg.Gtype = int32(pb.REDBLACK) //rb
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = rid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 百人场下注
func (rs *RoleActor) rbFreeBet(arg *pb.RBFreeBetReq, ctx actor.Context) {
	// if rs.User.IsTourist() {
	// 	rsp := new(pb.RBFreeBetRsp)
	// 	rsp.Error = pb.TouristInoperable
	// 	rs.Send(rsp)
	// 	return
	// }
	if rs.gamePid == nil {
		rsp := new(pb.RBFreeBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.RBFreeBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) rbStrategy(arg *pb.FreeSetRecord) {
	// if arg.Gtype != int32(Rpb.REDBLACK) {
	// 	return
	// }

	// if len(rs.RBDStrategy.RoundBet) > 0 {
	// 	lastBet := rs.RBDStrategy.RoundBet[len(rs.RBDStrategy.RoundBet)-1]
	// 	if arg.Bet >= lastBet*2 {
	// 		rs.RBDStrategy.LKYH.DoubleBetTimes++
	// 	}
	// }

	// rs.RBDStrategy.RoundBet = append(rs.RBDStrategy.RoundBet, arg.Bet)

	// if len(rs.RBDStrategy.RoundBet) >= 50 {
	// 	rs.RBDStrategy.RoundBet = rs.RBDStrategy.RoundBet[len(rs.RBDStrategy.RoundBet)-50:]
	// }
}
