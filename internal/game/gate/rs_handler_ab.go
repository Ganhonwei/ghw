package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) ABFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABFreeEnterRoomReq)
	glog.Debugf("ABFreeEnterRoomReq %#v", arg)
	rs.enterABFree(arg, ctx)
}

// 下注
func (rs *RoleActor) ABFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABFreeBetReq)
	// glog.Debugf("ABFreeBetReq %#v", arg)
	rs.ABdFreeBet(arg, ctx)
}

// 离开
func (rs *RoleActor) ABLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABLeaveReq)
	glog.Debugf("ABLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.ABLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 房间列表
func (rs *RoleActor) ABRoomListReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABRoomListReq)
	glog.Debugf("ABRoomListReq %#v", arg)
	rs.getABdRoomList(arg, ctx)
}

// 进入百人房间
func (rs *RoleActor) enterABFree(arg *pb.ABFreeEnterRoomReq, ctx actor.Context) {
	msg := rs.enterABMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterABMatchDesk(ctx actor.Context, roomid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.gamePid != nil {
		game2 := config.GetGame(rs.gameId)
		if game2.Gtype != int32(pb.ABAR) {
			rsp := new(pb.ABFreeEnterRoomRsp)
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
	msg.Name = cfg.Section("game.andarbahar").Name()
	msg.Gtype = int32(pb.ABAR) //andarbahar
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = roomid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getABdRoomList(arg *pb.ABRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.andarbahar").Name()
	msg.Gtype = int32(pb.ABAR) //7up
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
func (rs *RoleActor) ABdFreeBet(arg *pb.ABFreeBetReq, ctx actor.Context) {
	if rs.gamePid == nil {
		rsp := new(pb.ABFreeBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.ABFreeBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 创建私人房
func (rs *RoleActor) ABCreateRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ABCreateRoomReq)
	rsp := new(pb.ABCreateRoomRsp)
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return
	}

	// 基本配置充值要求
	pvpRoom := config.GetPvpRoom()
	if rs.User.GetMoney() < uint32(pvpRoom.PvpChargeLimit) { // && rs.RegistArea == 0 {
		rsp.Error = pb.PvpRechargeRequire
		rsp.RechargeRequire = uint32(pvpRoom.PvpChargeLimit)
		rs.Send(rsp)
		return
	}
	// arg.Gmode 0.真金 1.娱乐
	if arg.Gmode == 0 {
		game1 := config.GetGame(arg.GameId) // 选择的TP对战房
		if game1.Id == "" || game1.RoomType != 1 {
			glog.Errorf("ABCreateRoomReq gameId %s not found\n", arg.GameId)
			rsp.Error = pb.Failed
			rs.Send(rsp)
			return
		}
		score := rs.User.GetScore()
		// 庄家携带要求
		if score < int64(game1.AB.DealerAccess) {
			rsp.Error = pb.BeDealerNotEnough
			rs.Send(rsp)
			return
		}
		// 初始携带要求
		if score < int64(game1.MinFirstEntry) {
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
		// 每局最低准入
		if rs.User.GetScore() < int64(game1.Min_Access) {
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
	}

	if arg.Gmode == 1 {
		// 娱乐模式
		arg.GameId = "300"

		// 获取配置局数房费
		var cost int64
		pvp := config.GetPvpRoom()
		rounds := pvp.AbFunRoundCost[0]
		costs := pvp.AbFunRoundCost[1]
		for i, round := range rounds {
			if round == int(arg.Round) {
				cost = int64(costs[i])
			}
		}
		if cost > 0 && rs.GetDiamond() < cost {
			rsp.Error = pb.NotEnoughDiamond
			rs.Send(rsp)
			return
		}
	}
	msg := &pb.CreateDesk{
		Gmode:  arg.Gmode,
		Round:  arg.Round,
		GameId: arg.GameId,
	}
	msg.Name = cfg.Section("game.andarbahar").Name()
	msg.Gtype = int32(pb.ABAR)       //ABAR
	msg.Rtype = int32(pb.ROOM_TYPE1) //私人
	msg.Cid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//dbms中创建
	rs.dbmsPid.Request(msg, ctx.Self())
}

func (rs *RoleActor) ABEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABEnterRoomReq)
	glog.Debugf("ABEnterRoomReq %#v", arg)
	rs.enterABPriv(arg, ctx)
}

// 进入私人房间
func (rs *RoleActor) enterABPriv(arg *pb.ABEnterRoomReq, ctx actor.Context) {
	if rs.enterGame(ctx) {
		return
	}

	rsp := new(pb.ABEnterRoomRsp)

	// 被踢1分钟内不能再加入
	if utils.Timestamp()-5 < rs.User.BeKickoutTime {
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

	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	// 加入私人房间
	msg.Name = cfg.Section("game.andarbahar").Name()
	msg.Gtype = int32(pb.ABAR) //andarbahar
	msg.Dtype = int32(pb.DESK_TYPE_NORMAL)
	msg.Rtype = int32(pb.ROOM_TYPE1)
	msg.Code = arg.Code //邀请码
	rs.selectDesk(msg, ctx)
}

func (rs *RoleActor) ABKickoutRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABKickoutRoomReq)
	glog.Debugf("ABKickoutRoomReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABKickoutRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABChangeSeatReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABChangeSeatReq)
	glog.Debugf("ABChangeSeatReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABChangeSeatRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABChangeSeatAcceptReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABChangeSeatAcceptReq)
	glog.Debugf("ABChangeSeatAcceptReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABChangeSeatAcceptRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABLaunchVoteReq)
	glog.Debugf("ABLaunchVoteReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABLaunchVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABVoteReq)
	glog.Debugf("ABVoteReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABBetReq)
	glog.Debugf("ABBetReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ABGameStartReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ABGameStartReq)
	glog.Debugf("ABGameStartReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.ABGameStartRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}
