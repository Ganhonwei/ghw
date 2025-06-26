package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) LHFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHFreeEnterRoomReq)
	glog.Debugf("LHFreeEnterRoomReq %#v", arg)
	rs.enterLHFree(arg, ctx)
}

// 选择上庄，下庄
// func (rs *RoleActor) LHFreeDealerReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.LHFreeDealerReq)
// 	glog.Debugf("LHFreeDealerReq %#v", arg)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.LHFreeDealerRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 上庄列表
// func (rs *RoleActor) LHFreeDealerListReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.LHFreeDealerListReq)
// 	glog.Debugf("LHFreeDealerListReq %#v", arg)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.LHFreeDealerListRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 下注
func (rs *RoleActor) LHFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHFreeBetReq)
	// glog.Debugf("LHFreeBetReq %#v", arg)
	rs.lhdFreeBet(arg, ctx)
}

// 离开
func (rs *RoleActor) LHLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHLeaveReq)
	glog.Debugf("LHLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.LHLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入房间成功
func (rs *RoleActor) LHEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHEnterSuccessReq)
	glog.Debugf("LHEnterSuccessReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.LHEnterSuccessRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 房间列表
func (rs *RoleActor) LHRoomListReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHRoomListReq)
	glog.Debugf("LHRoomListReq %#v", arg)
	rs.getLhdRoomList(arg, ctx)
}

// 进入百人房间
func (rs *RoleActor) enterLHFree(arg *pb.LHFreeEnterRoomReq, ctx actor.Context) {
	msg := rs.enterLhdMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterLhdMatchDesk(ctx actor.Context, rid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.gamePid != nil {
		game2 := config.GetGame(rs.gameId)
		if game2.Gtype != int32(pb.LHD) {
			rsp := new(pb.LHFreeEnterRoomRsp)
			rsp.Error = pb.InOtherGame
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			rsp.GameType = game2.Gtype
			rsp.RoomType = int32(game2.RoomType)
			rs.Send(rsp)
			return nil
		}
	}
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.lhd").Name()
	msg.Gtype = int32(pb.LHD) //lhd
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = rid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getLhdRoomList(arg *pb.LHRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.lhd").Name()
	msg.Gtype = int32(pb.LHD) //lhd
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
func (rs *RoleActor) lhdFreeBet(arg *pb.LHFreeBetReq, ctx actor.Context) {
	// if rs.User.IsTourist() {
	// 	rsp := new(pb.LHFreeBetRsp)
	// 	rsp.Error = pb.TouristInoperable
	// 	rs.Send(rsp)
	// 	return
	// }
	if rs.gamePid == nil {
		rsp := new(pb.LHFreeBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.LHFreeBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) lhdStrategy(arg *pb.FreeSetRecord) {
	if arg.Gtype != int32(pb.LHD) {
		return
	}

	if len(rs.LHDStrategy.RoundBet) > 0 {
		lastBet := rs.LHDStrategy.RoundBet[len(rs.LHDStrategy.RoundBet)-1]
		if arg.Bet >= lastBet*2 {
			rs.LHDStrategy.LKYH.DoubleBetTimes++
		}
	}

	rs.LHDStrategy.RoundBet = append(rs.LHDStrategy.RoundBet, arg.Bet)

	if len(rs.LHDStrategy.RoundBet) >= 50 {
		rs.LHDStrategy.RoundBet = rs.LHDStrategy.RoundBet[len(rs.LHDStrategy.RoundBet)-50:]
	}
}

func (rs *RoleActor) LHDReset() {
	// 龙狂有祸
	rs.LHDStrategy.LKYH.HistoryT = append(rs.LHDStrategy.LKYH.HistoryT, rs.LHDStrategy.LKYH.TriggerTimes)
	if len(rs.LHDStrategy.LKYH.HistoryT) > 30 {
		rs.LHDStrategy.LKYH.HistoryT = rs.LHDStrategy.LKYH.HistoryT[len(rs.LHDStrategy.LKYH.HistoryT)-30:]
	}
	rs.LHDStrategy.LKYH.TriggerTimes = 0
	// 高潮涌现
	rs.LHDStrategy.GCYX.TriggerTimes = 0
}

// LHGCYXSync 高潮涌现状态同步
func (rs *RoleActor) LHGCYXSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.LHGCYXSync)

	rs.LHDStrategy.GCYX.Highing = arg.Highing
	rs.LHDStrategy.GCYX.TriggerTimes = int(arg.TriggerTimes)
	rs.LHDStrategy.GCYX.Hp = int(arg.Hp)
	rs.LHDStrategy.GCYX.C = int(arg.C)
	rs.LHDStrategy.GCYX.M = int(arg.M)
	rs.SevenStrategy.GCYX.BetAvg = arg.BetAvg
	if !rs.LHDStrategy.GCYX.Highing {
		// 高潮结束输赢清零
		rs.LHDStrategy.GCYX.Win = 0
		rs.LHDStrategy.GCYX.Lose = 0
	}
	rs.status = true
}

func (rs *RoleActor) CrashReset() {
	rs.CrashStrategy.GCYX.DailyGCTimes = 0
}

func (rs *RoleActor) AviatorReset() {
	rs.AvStrategy.GCYX.DailyGCTimes = 0
}
