package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) RMCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinEnterRoomReq)
	glog.Debugf("RMCoinEnterRoomReq %#v", arg)
	// arg.Id = ""
	rs.enterRMCoin(arg, ctx)
}

func (rs *RoleActor) RMFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeEnterRoomReq)
	glog.Debugf("RMFreeEnterRoomReq %#v", arg)
	// rs.enterRMFree(arg, ctx)
}

func (rs *RoleActor) RMFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeDealerReq)
	glog.Debugf("RMFreeDealerReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFreeDealerRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeDealerListReq)
	glog.Debugf("RMFreeDealerListReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFreeDealerListRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMSitReq)
	glog.Debugf("RMSitReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMSitRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeBetReq)
	glog.Debugf("RMFreeBetReq %#v", arg)
	// rs.nnRMFreeBet(arg, ctx)
}

func (rs *RoleActor) RMFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeTrendReq)
	glog.Debugf("RMFreeTrendReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFreeTrendRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeWinersReq)
	glog.Debugf("RMFreeWinersReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFreeWinersRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFreeRolesReq)
	glog.Debugf("RMFreeRolesReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFreeRolesRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMRoomListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMRoomListReq)
	glog.Debugf("RMRoomListReq %#v", arg)
	rs.getRMRoomList(arg, ctx)
}

func (rs *RoleActor) RMEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMEnterRoomReq)
	glog.Debugf("RMEnterRoomReq %#v", arg)
	rs.enterRMPriv(arg, ctx)
}

func (rs *RoleActor) RMCreateRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCreateRoomReq)
	glog.Debugf("RMCreateRoomReq %#v", arg)
	rs.createRMRoom(arg, ctx)
}

func (rs *RoleActor) RMLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMLeaveReq)
	glog.Debugf("RMLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMReadyReq(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.RMReadyReq)
	// glog.Debugf("RMReadyReq %#v", arg)
	// if rs.gamePid == nil {
	// 	rsp := new(pb.RMReadyRsp)
	// 	rsp.Error = pb.NotInRoom
	// 	rs.Send(rsp)
	// 	return
	// }
	// rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMReady2Req)
	glog.Debugf("RMReady2Req %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMReady2Rsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 手动开始游戏
func (rs *RoleActor) RMGameStartReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMGameStartReq)
	glog.Debugf("JHGameStartReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMGameStartRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMDrawCardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDrawCardReq)
	glog.Debugf("RMDrawCardReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMDrawCardRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMDiscardReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDiscardReq)
	glog.Debugf("RMDiscardReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMDiscardRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMFinishReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMFinishReq)
	glog.Debugf("RMFinishReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMFinishRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMDeclareReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDeclareReq)
	glog.Debugf("RMDeclareReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMDeclareRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMSortReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMSortReq)
	glog.Debugf("RMSortReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMSortRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMDropReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDropReq)
	glog.Debugf("RMDropReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMDropRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMDropScoreReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMDropScoreReq)
	glog.Debugf("RMDropScoreReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMDropScoreRsp)
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

func (rs *RoleActor) RMGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMGameRecordReq)
	glog.Debugf("RMGameRecordReq %#v", arg)
	//TODO
}

func (rs *RoleActor) RMLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMLaunchVoteReq)
	glog.Debugf("RMLaunchVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMLaunchVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMVoteReq)
	glog.Debugf("RMVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMAutoSortReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMAutoSortReq)
	glog.Debugf("RMAutoSortReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMAutoSortRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMQiCardsReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMQiCardsReq)
	glog.Debugf("RMQiCardsReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMQiCardsRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinSeeReq)
	glog.Debugf("RMCoinSeeReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinSeeRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinCallReq)
	glog.Debugf("RMCoinCallReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinCallRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinRaiseReq)
	glog.Debugf("RMCoinRaiseReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinRaiseRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinFoldReq)
	glog.Debugf("RMCoinFoldReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinFoldRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinBiReq)
	glog.Debugf("RMCoinBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinReplyBiReq)
	glog.Debugf("RMCoinReplyBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinReplyBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) RMCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RMCoinChangeRoomReq)
	glog.Debugf("RMCoinChangeRoomReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.RMCoinChangeRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入百人房间
// func (rs *RoleActor) enterRMFree(arg *pb.RMFreeEnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterRMMatchDesk(ctx)
// 	if msg != nil {
// 		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入私人房间
func (rs *RoleActor) enterRMPriv(arg *pb.RMEnterRoomReq, ctx actor.Context) {
	msg := rs.enterRMMatchDesk(ctx)
	if msg == nil {
		return
	}

	rsp := new(pb.JHEnterRoomRsp)

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

	msg.Rtype = int32(pb.ROOM_TYPE1) //私人
	msg.Code = arg.Code              //邀请码
	rs.selectDesk(msg, ctx)
}

// 进入自由房间
func (rs *RoleActor) enterRMCoin(arg *pb.RMCoinEnterRoomReq, ctx actor.Context) {
	game1 := config.GetGame(arg.Gameid)
	if rs.gamePid != nil {
		game2 := config.GetGame(rs.gameId)
		if game1.Gtype != game2.Gtype { //想进的游戏和已经在的游戏不同
			rsp := new(pb.RMCoinEnterRoomRsp)
			rsp.Error = pb.InOtherGame
			rsp.Gameid = rs.gameId
			rsp.Roomid = rs.roomId
			rsp.GameType = game2.Gtype
			rsp.RoomType = int32(game2.RoomType)
			rs.Send(rsp)
			return
		}
	}
	var msg *pb.MatchDesk
	switch game1.Gtype {
	case int32(pb.RUMMY2): // rm双人
		msg = rs.enterRM2MatchDesk(ctx)
	default:
		msg = rs.enterRMMatchDesk(ctx)
	}

	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE0) //自由
		msg.Gameid = arg.Gameid          //游戏ID
		msg.Roomid = arg.Roomid          //房间ID
		// msg.Dtype = int32(pb.DESK_TYPE1) //玩法类型
		//计算出匹配房间等级,算法一致
		//msg.Ltype = int32(pb.ROOM_LEVEL1) //等级
		msg.Ltype = handler.MatchLevel(rs.User.GetCoin())
		if msg.Ltype < 0 {
			rsp := new(pb.RMCoinEnterRoomRsp)
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
		if !rs.Robot {
			//判断最低进入
			game := config.GetGame(arg.Gameid)
			if rs.User.GetScore() < int64(game.Min_Access) {
				rsp := new(pb.RMCoinEnterRoomRsp)
				rsp.Error = pb.NotEnoughCoin
				rs.Send(rsp)
				return
			}
			//判断最高进入
			if game.Max_Access != -1 && rs.User.GetScore() > int64(game.Max_Access) {
				rsp := new(pb.RMCoinEnterRoomRsp)
				rsp.Error = pb.TooManyCoin
				rs.Send(rsp)
				return
			}
		}

		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterRMMatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rummy").Name()
	msg.Gtype = int32(pb.RUMMY) //rummy
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) enterRM2MatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rummy_2").Name()
	msg.Gtype = int32(pb.RUMMY2) //rummy2
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getRMRoomList(arg *pb.RMRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rm").Name()
	msg.Gtype = int32(pb.RUMMY) //金花
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
// func (rs *RoleActor) nnRMFreeBet(arg *pb.RMFreeBetReq, ctx actor.Context) {
// 	if rs.User.IsTourist() {
// 		rsp := new(pb.RMFreeBetRsp)
// 		rsp.Error = pb.TouristInoperable
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.gamePid == nil {
// 		rsp := new(pb.RMFreeBetRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	value := arg.GetValue()
// 	seat := arg.GetSeat()
// 	if !(seat >= uint32(pb.DESK_SEAT2) &&
// 		seat <= uint32(pb.DESK_SEAT9)) {
// 		rsp := new(pb.RMFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if value <= 0 {
// 		rsp := new(pb.RMFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.User.GetCoin() < int64(value) {
// 		rsp := new(pb.RMFreeBetRsp)
// 		rsp.Error = pb.NotEnoughCoin
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 创建房间
func (rs *RoleActor) createRMRoom(arg *pb.RMCreateRoomReq, ctx actor.Context) {
	rsp := new(pb.RMCreateRoomRsp)
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
		if game1.Id == "" {
			glog.Errorf("RMCreateRoomReq gameId %s not found\n", arg.GameId)
			rsp.Error = pb.Failed
			rs.Send(rsp)
			return
		}
		score := rs.User.GetScore()
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
		arg.GameId = "200"

		// 获取配置局数房费
		var cost int64
		pvp := config.GetPvpRoom()
		rounds := pvp.RmFunRoundCost[0]
		costs := pvp.RmFunRoundCost[1]
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
	msg.Name = cfg.Section("game.rummy").Name()
	msg.Gtype = int32(pb.RUMMY)      //RM
	msg.Rtype = int32(pb.ROOM_TYPE1) //私人
	msg.Cid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//dbms中创建
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 私人房踢人
func (rs *RoleActor) RMKickoutRoomReq(ctx actor.Context) {
	if rs.gamePid == nil {
		rsp := new(pb.RMKickoutRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}

	msg := ctx.Message().(*pb.RMKickoutRoomReq)
	rs.gamePid.Request(msg, ctx.Self())
}

// 私人房换座
func (rs *RoleActor) RMChangeSeatReq(ctx actor.Context) {
	if rs.gamePid == nil {
		rsp := new(pb.RMChangeSeatRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}

	msg := ctx.Message().(*pb.RMChangeSeatReq)
	rs.gamePid.Request(msg, ctx.Self())
}

func (rs *RoleActor) RMRoiRecordsync(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMRoiRecordsync)

	if msg.RoiId != "" {
		if rs.User.RmRoiDayLimits == nil {
			rs.User.RmRoiDayLimits = make(map[string]int32)
		}
		if rs.User.RmRoiLimits == nil {
			rs.User.RmRoiLimits = make(map[string]int32)
		}
		rs.User.RmRoiDayLimits[msg.RoiId]++
		rs.User.RmRoiLimits[msg.RoiId]++
	}

	rs.User.RmWithout1stDrop = int(msg.RmWithout1StDrop)
	rs.User.RmWithout1stNotDrop = int(msg.RmWithout1StNotDrop)
}
