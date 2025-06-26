package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (rs *RoleActor) JHCoinEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinEnterRoomReq)
	glog.Debugf("JHCoinEnterRoomReq %#v", arg)
	// arg.Id = ""
	rs.enterJHCoin(arg, ctx)
}

func (rs *RoleActor) JHFreeEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeEnterRoomReq)
	glog.Debugf("JHFreeEnterRoomReq %#v", arg)
	// rs.enterJHFree(arg, ctx)
}

func (rs *RoleActor) JHFreeDealerReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeDealerReq)
	glog.Debugf("JHFreeDealerReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHFreeDealerRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHFreeDealerListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeDealerListReq)
	glog.Debugf("JHFreeDealerListReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHFreeDealerListRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHSitReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHSitReq)
	glog.Debugf("JHSitReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHSitRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHFreeBetReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeBetReq)
	glog.Debugf("JHFreeBetReq %#v", arg)
	// rs.nnJHFreeBet(arg, ctx)
}

func (rs *RoleActor) JHFreeTrendReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeTrendReq)
	glog.Debugf("JHFreeTrendReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHFreeTrendRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHFreeWinersReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeWinersReq)
	glog.Debugf("JHFreeWinersReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHFreeWinersRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHFreeRolesReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHFreeRolesReq)
	glog.Debugf("JHFreeRolesReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHFreeRolesRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHRoomListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHRoomListReq)
	glog.Debugf("JHRoomListReq %#v", arg)
	rs.getJHRoomList(arg, ctx)
}

func (rs *RoleActor) JHEnterRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHEnterRoomReq)
	glog.Debugf("JHEnterRoomReq %#v", arg)
	rs.enterJHPriv(arg, ctx)
}

func (rs *RoleActor) JHCreateRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCreateRoomReq)
	glog.Debugf("JHCreateRoomReq %#v", arg)
	rs.createJHRoom(arg, ctx)
}

func (rs *RoleActor) JHLeaveReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHLeaveReq)
	glog.Debugf("JHLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHReadyReq(ctx actor.Context) {
	// msg := ctx.Message()
	// arg := msg.(*pb.JHReadyReq)
	// glog.Debugf("JHReadyReq %#v", arg)
	// if rs.gamePid == nil {
	// 	rsp := new(pb.JHReadyRsp)
	// 	rsp.Error = pb.NotInRoom
	// 	rs.Send(rsp)
	// 	return
	// }
	// rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHReady2Req(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHReady2Req)
	glog.Debugf("JHReady2Req %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHReady2Rsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHGameStartReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHGameStartReq)
	glog.Debugf("JHGameStartReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHGameStartRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) ChatTextReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChatTextReq)
	glog.Debugf("ChatTextReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.ChatTextRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	// 付费
	// vip := table.GetTables().VipTable.Get(int32(rs.Vip.Lv))
	// if vip.EmojiCost < 0 {
	// 	rsp := new(pb.ChatTextRsp)
	// 	rsp.Error = pb.VipLvTooLow
	// 	rs.Send(rsp)
	// 	return
	// }
	// if int64(vip.EmojiCost) > rs.Diamond {
	// 	rsp := new(pb.ChatTextRsp)
	// 	rsp.Error = pb.NotEnoughCoin
	// 	rs.Send(rsp)
	// 	return
	// }
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) HedgeTrigger(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.HedgeTrigger)
	glog.Debugf("HedgeTrigger %#v", arg)
	rs.Hedge = rs.Hedge<<1 | 1
	rs.status = true
}

func (rs *RoleActor) JHGameRecordReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHGameRecordReq)
	glog.Debugf("JHGameRecordReq %#v", arg)
	//TODO
}

func (rs *RoleActor) JHLaunchVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHLaunchVoteReq)
	glog.Debugf("JHLaunchVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHLaunchVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHVoteReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHVoteReq)
	glog.Debugf("JHVoteReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHVoteRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinSeeReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinSeeReq)
	glog.Debugf("JHCoinSeeReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinSeeRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinCallReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinCallReq)
	glog.Debugf("JHCoinCallReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinCallRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinRaiseReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinRaiseReq)
	glog.Debugf("JHCoinRaiseReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinRaiseRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinFoldReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinFoldReq)
	glog.Debugf("JHCoinFoldReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinFoldRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinBiReq)
	glog.Debugf("JHCoinBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinReplyBiReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinReplyBiReq)
	glog.Debugf("JHCoinReplyBiReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinReplyBiRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) JHCoinChangeRoomReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JHCoinChangeRoomReq)
	glog.Debugf("JHCoinChangeRoomReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.JHCoinChangeRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入百人房间
// func (rs *RoleActor) enterJHFree(arg *pb.JHFreeEnterRoomReq, ctx actor.Context) {
// 	msg := rs.enterJHMatchDesk(ctx)
// 	if msg != nil {
// 		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
// 		rs.selectDesk(msg, ctx)
// 	}
// }

// 进入私人房间
func (rs *RoleActor) enterJHPriv(arg *pb.JHEnterRoomReq, ctx actor.Context) {
	msg := rs.enterJHMatchDesk(ctx)
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

	// 加入私人房间
	msg.Rtype = int32(pb.ROOM_TYPE1)
	msg.Code = arg.Code //邀请码
	rs.selectDesk(msg, ctx)
}

// 进入自由房间
func (rs *RoleActor) enterJHCoin(arg *pb.JHCoinEnterRoomReq, ctx actor.Context) {
	game1 := config.GetGame(arg.Gameid)
	if rs.gamePid != nil {
		game1 := config.GetGame(arg.Gameid)
		game2 := config.GetGame(rs.gameId)
		if game1.Gtype != game2.Gtype { //想进的游戏和已经在的游戏不同
			rsp := new(pb.JHCoinEnterRoomRsp)
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
	case int32(pb.HUA2): // sa服tp
		msg = rs.enterJH2MatchDesk(ctx)
	default:
		msg = rs.enterJHMatchDesk(ctx)
	}

	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE0)      //自由
		msg.Gameid = arg.Gameid               //游戏ID
		msg.Roomid = arg.Roomid               //房间ID
		msg.IsChangeTable = arg.IsChangeTable // 是否是换桌
		// msg.Dtype = int32(pb.DESK_TYPE1) //玩法类型
		//计算出匹配房间等级,算法一致
		//msg.Ltype = int32(pb.ROOM_LEVEL1) //等级
		msg.Ltype = handler.MatchLevel(rs.User.GetCoin())
		if msg.Ltype < 0 {
			rsp := new(pb.JHCoinEnterRoomRsp)
			rsp.Error = pb.NotEnoughCoin
			rs.Send(rsp)
			return
		}
		if !rs.Robot && !rs.SimRobot {
			//判断最低进入
			game := config.GetGame(arg.Gameid)

			if arg.IsChangeTable {
				if rs.User.GetScore() < int64(game.Kick_Score) {
					rsp := new(pb.JHCoinEnterRoomRsp)
					rsp.Error = pb.NotEnoughCoin
					rs.Send(rsp)
					return
				}
			} else {
				if rs.User.GetScore() < int64(game.Min_Access) {
					rsp := new(pb.JHCoinEnterRoomRsp)
					rsp.Error = pb.NotEnoughCoin
					rs.Send(rsp)
					return
				}
			}

			//判断最高进入
			if game.Max_Access != -1 && rs.User.GetScore() > int64(game.Max_Access) {
				rsp := new(pb.JHCoinEnterRoomRsp)
				rsp.Error = pb.TooManyCoin
				rs.Send(rsp)
				return
			}
		}

		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterJHMatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.hua").Name()
	msg.Gtype = int32(pb.HUA) //金花
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) enterJH2MatchDesk(ctx actor.Context) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.hua_2").Name()
	msg.Gtype = int32(pb.HUA2) //金花
	msg.Rtype = int32(pb.ROOM_TYPE0)
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getJHRoomList(arg *pb.JHRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.hua").Name()
	msg.Gtype = int32(pb.HUA) //金花
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
// func (rs *RoleActor) nnJHFreeBet(arg *pb.JHFreeBetReq, ctx actor.Context) {
// 	if rs.User.IsTourist() {
// 		rsp := new(pb.JHFreeBetRsp)
// 		rsp.Error = pb.TouristInoperable
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.gamePid == nil {
// 		rsp := new(pb.JHFreeBetRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	value := arg.GetValue()
// 	seat := arg.GetSeat()
// 	if !(seat >= uint32(pb.DESK_SEAT2) &&
// 		seat <= uint32(pb.DESK_SEAT9)) {
// 		rsp := new(pb.JHFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if value <= 0 {
// 		rsp := new(pb.JHFreeBetRsp)
// 		rsp.Error = pb.OperateError
// 		rs.Send(rsp)
// 		return
// 	}
// 	if rs.User.GetCoin() < int64(value) {
// 		rsp := new(pb.JHFreeBetRsp)
// 		rsp.Error = pb.NotEnoughCoin
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 创建房间
func (rs *RoleActor) createJHRoom(arg *pb.JHCreateRoomReq, ctx actor.Context) {
	rsp := new(pb.JHCreateRoomRsp)
	// if rs.gamePid != nil {
	// 	game2 := config.GetGame(rs.gameId)
	// 	// if game2.Gtype != int32(pb.HUA) {
	// 	rsp := new(pb.JHCoinEnterRoomRsp)
	// 	rsp.Error = pb.InOtherGame
	// 	rsp.Gameid = rs.gameId
	// 	rsp.Roomid = rs.roomId
	// 	rsp.GameType = game2.Gtype
	// 	rsp.RoomType = int32(game2.RoomType)
	// 	rs.Send(rsp)
	// 	return
	// 	// }
	// }
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
			glog.Errorf("JHCreateRoomReq gameId %s not found\n", arg.GameId)
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
		arg.GameId = "100"

		// 获取配置局数房费
		var cost int64
		pvp := config.GetPvpRoom()
		rounds := pvp.TpFunRoundCost[0]
		costs := pvp.TpFunRoundCost[1]
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
	msg.Name = cfg.Section("game.hua").Name()
	msg.Gtype = int32(pb.HUA)        //TP
	msg.Rtype = int32(pb.ROOM_TYPE1) //私人
	msg.Cid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//dbms中创建
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 自由房踢人
func (rs *RoleActor) JHKickoutRoomReq(ctx actor.Context) {
	if rs.gamePid == nil {
		rsp := new(pb.JHKickoutRoomRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}

	msg := ctx.Message().(*pb.JHKickoutRoomReq)
	rs.gamePid.Request(msg, ctx.Self())
}

// 自由房换座位
func (rs *RoleActor) JHChangeSeatReq(ctx actor.Context) {
	if rs.gamePid == nil {
		rsp := new(pb.JHChangeSeatRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}

	msg := ctx.Message().(*pb.JHChangeSeatReq)
	rs.gamePid.Request(msg, ctx.Self())
}

// TP新手试探局数据游戏中同步到网关
func (rs *RoleActor) JHSetTpNewbieProbe(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHSetTpNewbieProbe)
	rs.User.TpNewbieProbeId = msg.TpNewbieProbeId
	rs.User.TpNewbieStableId = msg.TpNewbieStableId
	rs.User.TpNewbieStableRounds = msg.TpNewbieStableRounds
}

// TP偷鸡被偷鸡输赢分同步到网关
func (rs *RoleActor) JHSetTpHurtScore(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHSetTpHurtScore)
	rs.User.TpHurtWinScore += int64(msg.HurtScore)
	rs.User.TpBeHurtLoseScore += int64(msg.BeHurtScore)
	if msg.HurtScore != 0 {
		rs.User.TpHurtWinRound++
	}
	if msg.BeHurtScore != 0 {
		rs.User.TpBeHurtLoseRound++
	}
	rs.status = true
}

// TP vh更新同步到网关
func (rs *RoleActor) JHVHRecordSync(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHVHRecordSync)
	if rs.TpUserVHRecord == nil {
		rs.TpUserVHRecord = make(map[int32][6]int64)
	}
	rs.TpUserVHRecord[msg.VhKey] = [6]int64(msg.VhRecord)
	rs.status = true
}

// TP 高牌换牌心跳同步网关
func (rs *RoleActor) JHHeartbeatSync(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHHeartbeatSync)
	rs.TpHeartbeatRound++
	if msg.AddRound {
		rs.TpHighCardRounds++
	}
	if msg.ResetRound {
		rs.TpHighCardRounds = 0
		rs.TpHeartbeatResetTimes++
	}
	rs.status = true
}

// TP 乐极生悲被冤局CD
func (rs *RoleActor) JHUserLjsbPyCDSync(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHUserLjsbPyCDSync)
	rs.TpUserLjsbPyCD = msg.PyCD
	rs.status = true
}

// TP 高潮涌现同步
func (rs *RoleActor) JHUserGCYXSync(ctx actor.Context) {
	msg := ctx.Message().(*pb.JHUserGCYXSync)
	if msg.TpUserGcyxCD != 0 {
		rs.TpUserGcyxCD = msg.TpUserGcyxCD
	}
	if msg.TpUserGcyxHp != 0 {
		rs.TpUserGcyxHp = msg.TpUserGcyxHp
	}
	if msg.TpUserGcyxWin != 0 {
		rs.TpUserGcyxWin += msg.TpUserGcyxWin
	}
	if msg.TpUserGcyxLose != 0 {
		rs.TpUserGcyxLose += msg.TpUserGcyxLose
	}
	rs.status = true
}
