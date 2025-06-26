package gate

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data/ck"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) UPFreeEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPFreeEnterRoomReq)
	glog.Debugf("UPFreeEnterRoomReq %#v", arg)
	rs.enterUPFree(arg, ctx)
}

// 进入房间成功
func (rs *RoleActor) UPEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPEnterSuccessReq)
	glog.Debugf("UPEnterSuccessReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.UPEnterSuccessRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 玩家倍投下注(当前下注每一门注码都加倍)
func (rs *RoleActor) UPDoubleBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPDoubleBetReq)
	glog.Debugf("UPDoubleBetReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.UPDoubleBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 撤销上一步下注
func (rs *RoleActor) UPUndoBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPUndoBetReq)
	glog.Debugf("UPUndoBetReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.UPUndoBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 下注
func (rs *RoleActor) UPFreeBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPFreeBetReq)
	// glog.Debugf("UPFreeBetReq %#v", arg)
	rs.UPdFreeBet(arg, ctx)
}

// 离开
func (rs *RoleActor) UPLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPLeaveReq)
	glog.Debugf("UPLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.UPLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 房间列表
func (rs *RoleActor) UPRoomListReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPRoomListReq)
	glog.Debugf("UPRoomListReq %#v", arg)
	rs.getUPdRoomList(arg, ctx)
}

// 进入百人房间
func (rs *RoleActor) enterUPFree(arg *pb.UPFreeEnterRoomReq, ctx actor.Context) {
	msg := rs.enterUPMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterUPMatchDesk(ctx actor.Context, roomid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.gamePid != nil {
		// game2 := config.GetGame(rs.gameId)
		// if game2.Gtype != int32(pb.SEVEN) {
		if rs.gtype != int32(pb.SEVEN) {
			rsp := new(pb.UPFreeEnterRoomRsp)
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
	msg.Name = cfg.Section("game.7up").Name()
	msg.Gtype = int32(pb.SEVEN) //UPd
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = roomid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getUPdRoomList(arg *pb.UPRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.7up").Name()
	msg.Gtype = int32(pb.SEVEN) //7up
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 百人场下注
func (rs *RoleActor) UPdFreeBet(arg *pb.UPFreeBetReq, ctx actor.Context) {
	// if rs.User.IsTourist() {
	// 	rsp := new(pb.UPFreeBetRsp)
	// 	rsp.Error = pb.TouristInoperable
	// 	rs.Send(rsp)
	// 	return
	// }
	if rs.gamePid == nil {
		rsp := new(pb.UPFreeBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.UPFreeBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

func (rs *RoleActor) sevenUpStrategy(arg *pb.FreeSetRecord) {
	if arg.Gtype != int32(pb.SEVEN) {
		return
	}

	rs.SevenStrategy.RoundBet = append(rs.SevenStrategy.RoundBet, arg.Bet)
	if len(rs.SevenStrategy.RoundBet) >= 50 {
		rs.SevenStrategy.RoundBet = rs.SevenStrategy.RoundBet[len(rs.SevenStrategy.RoundBet)-50:]
	}
}

func (rs *RoleActor) UPReset() {
	// 7管严
	rs.SevenStrategy.LKYH.HistoryT = append(rs.SevenStrategy.LKYH.HistoryT, rs.SevenStrategy.LKYH.TriggerTimes)
	if len(rs.SevenStrategy.LKYH.HistoryT) > 30 {
		rs.SevenStrategy.LKYH.HistoryT = rs.SevenStrategy.LKYH.HistoryT[len(rs.SevenStrategy.LKYH.HistoryT)-30:]
	}
	rs.SevenStrategy.LKYH.TriggerTimes = 0
	// 高潮涌现
	rs.LHDStrategy.GCYX.TriggerTimes = 0
}

// UPGCYXSync 高潮涌现状态同步
func (rs *RoleActor) UPGCYXSync(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPGCYXSync)

	rs.SevenStrategy.GCYX.Highing = arg.Highing
	rs.SevenStrategy.GCYX.TriggerTimes = int(arg.TriggerTimes)
	rs.SevenStrategy.GCYX.Hp = int(arg.Hp)
	rs.SevenStrategy.GCYX.C = int(arg.C)
	rs.SevenStrategy.GCYX.M = int(arg.M)
	rs.SevenStrategy.GCYX.BetAvg = arg.BetAvg
	if !rs.SevenStrategy.GCYX.Highing {
		// 高潮结束输赢清零
		rs.SevenStrategy.GCYX.Win = 0
		rs.SevenStrategy.GCYX.Lose = 0
	}
	rs.status = true
}

// 7up 个人记录查询
func (rs *RoleActor) UPMyHistoryReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.UPMyHistoryReq)
	rsp := new(pb.UPMyHistoryRsp)
	defer rs.Send(rsp)

	userid := rs.Userid
	pageSize := arg.PageSize
	if pageSize <= 0 {
		pageSize = 100
	} else if pageSize > 200 {
		pageSize = 200
	}

	if arg.Date == "" {
		arg.Date = time.Now().In(location).Format(utils.FORMAT_DATE)
	}
	// 筛选日期
	sdate := fmt.Sprintf("%s 00:00:00", arg.Date)
	stime, err := time.ParseInLocation(utils.FORMAT, sdate, location)
	if err != nil {
		rsp.Error = pb.Failed
		return
	}
	etime := stime.AddDate(0, 0, 1)

	args := []any{userid, stime.Unix(), etime.Unix()}

	cond := ""
	if arg.PrevLastId != "" {
		cond += " AND begin_time < (SELECT begin_time FROM game.col_detail FINAL WHERE id = ? AND userid = ? AND gtype = 3 LIMIT 1)"
		args = append(args, arg.PrevLastId, userid)
	}
	switch arg.Change {
	case 1:
		cond += " AND score > 0"
	case 2:
	}

	sql := `
		SELECT id, begin_time, end_time, win_type, score, bet_amount
		FROM game.col_detail FINAL
		WHERE userid = ? AND begin_time >= ? AND begin_time < ? AND robot = 0 AND gtype = 3 AND win_type in (1,2,3) %s
		ORDER BY begin_time DESC
		LIMIT ?
	`
	args = append(args, pageSize)

	var datas []map[string]any
	err = ck.Select(&datas, fmt.Sprintf(sql, cond), args...)
	if err != nil {
		rsp.Error = pb.Failed
		return
	}
	for _, data := range datas {
		id := data["id"].(string)
		begin_time := utils.ToInt64(data["begin_time"])
		end_time := utils.ToInt64(data["end_time"])
		win_type := utils.ToInt64(data["win_type"])
		score := utils.ToInt64(data["score"])
		bet_amount := utils.ToInt64(data["bet_amount"])

		beginTime := time.Unix(begin_time, 0).In(location)
		endTime := time.Unix(end_time, 0).In(location)

		record := &pb.UPMyHistory{
			Id:      id,
			Sdate:   beginTime.Format(utils.FORMAT_DATE),
			Stime:   beginTime.Format(utils.FORMAT_TIME),
			Edate:   endTime.Format(utils.FORMAT_DATE),
			Etime:   endTime.Format(utils.FORMAT_TIME),
			Bets:    bet_amount,
			Score:   utils.CaseElse(score > 0, score+bet_amount, 0),
			WinType: int32(win_type),
		}
		rsp.Records = append(rsp.Records, record)
	}
	rsp.HasMore = len(rsp.Records) >= int(arg.PageSize)
}
