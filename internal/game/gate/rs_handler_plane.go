package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) PLANEEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEEnterRoomReq)
	glog.Debugf("PLANEEnterRoomReq %#v", arg)
	msg := rs.enterPlaneMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 下注
func (rs *RoleActor) PLANEBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEBetReq)
	if rs.gamePid == nil {
		rsp := new(pb.PLANEBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.PLANEBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 取消下注
func (rs *RoleActor) PLANECancelBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANECancelBetReq)
	if rs.gamePid == nil {
		rsp := new(pb.PLANECancelBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	arg.Userid = rs.Userid
	rs.gamePid.Request(arg, ctx.Self())
}

// 撤离
func (rs *RoleActor) PLANEBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEBackReq)
	if rs.gamePid == nil {
		rsp := new(pb.PLANEBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 房间列表
func (rs *RoleActor) PLANERoomListReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANERoomListReq)
	glog.Debugf("PLANERoomListReq %#v", arg)
	rs.getPlaneRoomList(arg, ctx)
}

// 设置自动撤离
func (rs *RoleActor) PLANECrashMultipleReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANECrashMultipleReq)
	glog.Debugf("PLANECrashMultipleReq %#v", arg)
	rs.autoPlaneMultiple(arg, ctx)
}

// 上一局记录
func (rs *RoleActor) PLANELastRoundRecordReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANELastRoundRecordReq)
	glog.Debugf("PLANELastRoundRecordReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.PLANELastRoundRecordRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 离开
func (rs *RoleActor) PLANELeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANELeaveReq)
	glog.Debugf("PLANELeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.PLANELeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入或匹配桌子
func (rs *RoleActor) enterPlaneMatchDesk(ctx actor.Context, rid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.gamePid != nil {
		// game2 := config.GetGame(rs.gameId)
		// if game2.Gtype != int32(pb.PLANE) {
		if rs.gtype != int32(pb.PLANE) {
			rsp := new(pb.PLANEEnterRoomRsp)
			rsp.Error = pb.InOtherGame
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			rsp.GameType = rs.gtype
			rsp.RoomType = 0
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
	msg.Name = cfg.Section("game.plane").Name()
	msg.Gtype = int32(pb.PLANE) //plane
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = rid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getPlaneRoomList(arg *pb.PLANERoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.plane").Name()
	msg.Gtype = int32(pb.PLANE) //plane
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 设置自动撤离
func (rs *RoleActor) autoPlaneMultiple(arg *pb.PLANECrashMultipleReq, ctx actor.Context) {
	user := rs.User
	user.PlaneMultiple = arg.Multiple
	user.PlaneAutoLeave = arg.AutoCrash
	rs.status = true
	rsp := &pb.PLANECrashMultipleRsp{
		Multiple:  arg.Multiple,
		AutoCrash: arg.AutoCrash,
	}
	rs.Send(rsp)
	if rs.gamePid != nil {
		rs.gamePid.Request(arg, ctx.Self())
	}
}

// plane 个人记录查询
func (rs *RoleActor) PLANEMyHistoryReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANEMyHistoryReq)

	arg2 := &pb.CrashMyHistoryReq{Date: arg.Date, PrevLastId: arg.PrevLastId, PageSize: arg.PageSize}
	rsp2 := rs.crashMyHistoryReq(arg2, int32(pb.PLANE))

	rsp := &pb.PLANEMyHistoryRsp{
		Error:   rsp2.Error,
		Date:    rsp2.Date,
		Records: rsp2.Records,
		HasMore: rsp2.HasMore,
	}
	rs.Send(rsp)
}

// plane 大赢家记录查询
func (rs *RoleActor) PLANETopWinnersReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PLANETopWinnersReq)

	arg2 := &pb.CrashTopWinnersReq{Range: arg.Range}
	rsp2 := rs.crashTopWinnersReq(arg2, int32(pb.PLANE))

	rsp := &pb.PLANETopWinnersRsp{
		Error:     rsp2.Error,
		DateRange: rsp2.DateRange,
		Records:   rsp2.Records,
	}
	rs.Send(rsp)
}
