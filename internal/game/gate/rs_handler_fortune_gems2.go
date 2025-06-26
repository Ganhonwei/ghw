package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) FortuneGems2EnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGems2EnterRoomReq)
	glog.Debugf("FortuneGems2EnterRoomReq %#v", arg)
	rs.enterFortuneGems2(arg, ctx)
}

// 下注
func (rs *RoleActor) FortuneGems2BetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGems2BetReq)
	// glog.Debugf("RBFreeBetReq %#v", arg)

	if rs.gamePid == nil {
		rsp := new(pb.FortuneGems2BetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}

	if arg.Bet <= 0 {
		rsp := new(pb.FortuneGems2BetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 离开
func (rs *RoleActor) FortuneGems2LeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.FortuneGems2LeaveReq)
	glog.Debugf("FortuneGems2LeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.FortuneGems2LeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入fortune_gems2房间
func (rs *RoleActor) enterFortuneGems2(arg *pb.FortuneGems2EnterRoomReq, ctx actor.Context) {
	msg := rs.enterFortuneGems2MatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE0) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterFortuneGems2MatchDesk(ctx actor.Context, rid string) *pb.MatchDesk {
	// 已经在游戏中,直接加入
	if rs.enterGame(ctx) {
		return nil
	}

	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.fortune_gems2").Name()
	msg.Gameid = pb.GameType_name[int32(pb.FORTUNE_GEMS2)]
	msg.Gtype = int32(pb.FORTUNE_GEMS2) // fortune_gems2
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE0)
	msg.Roomid = rid
	return msg
}

// 踩雷
// func (rs *RoleActor) FortuneGems2PitStepReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGems2PitStepReq)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.FortuneGems2PitStepRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 提现
// func (rs *RoleActor) FortuneGems2CashOutReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGems2CashOutReq)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.FortuneGems2CashOutRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// 取消自动下注
// func (rs *RoleActor) FortuneGems2AutoBetCancelReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGems2AutoBetCancelReq)
// 	if rs.gamePid == nil {
// 		rsp := new(pb.FortuneGems2AutoBetCancelRsp)
// 		rsp.Error = pb.NotInRoom
// 		rs.Send(rsp)
// 		return
// 	}
// 	rs.gamePid.Request(arg, ctx.Self())
// }

// crash 个人记录查询
// func (rs *RoleActor) FortuneGems2MyHistoryReq(ctx actor.Context) {
// 	arg := ctx.Message().(*pb.FortuneGems2MyHistoryReq)
// 	rsp := new(pb.FortuneGems2MyHistoryRsp)
// 	defer rs.Send(rsp)

// 	userid := rs.Userid
// 	pageSize := arg.PageSize
// 	if pageSize <= 0 {
// 		pageSize = 100
// 	} else if pageSize > 200 {
// 		pageSize = 200
// 	}

// 	if arg.Date == "" {
// 		arg.Date = time.Now().In(location).Format(utils.FORMAT_DATE)
// 	}
// 	// 筛选日期
// 	sdate := fmt.Sprintf("%s 00:00:00", arg.Date)
// 	stime, err := time.ParseInLocation(utils.FORMAT, sdate, location)
// 	if err != nil {
// 		rsp.Error = pb.Failed
// 		return
// 	}
// 	etime := stime.AddDate(0, 0, 1)

// 	args := []any{userid, stime.Unix(), etime.Unix()}

// 	cond := ""
// 	if arg.PrevLastId != "" {
// 		cond += " AND begin_time < (SELECT begin_time FROM game.col_detail FINAL WHERE id = ? AND userid = ? AND gtype = 14 LIMIT 1)"
// 		args = append(args, arg.PrevLastId, userid)
// 	}
// 	switch arg.Change {
// 	case 1:
// 		cond += " AND score > 0"
// 	case 2:
// 	}

// 	sql := `
// 		SELECT id, begin_time, end_time, win_type, score, bet_amount, mines_multiple
// 		FROM game.col_detail FINAL
// 		WHERE userid = ? AND begin_time >= ? AND begin_time < ? AND robot = 0 AND gtype = 14 AND win_type in (1,2,3) %s
// 		ORDER BY begin_time DESC
// 		LIMIT ?
// 	`
// 	args = append(args, pageSize)

// 	var datas []map[string]any
// 	err = ck.Select(&datas, fmt.Sprintf(sql, cond), args...)
// 	if err != nil {
// 		rsp.Error = pb.Failed
// 		return
// 	}
// 	for _, data := range datas {
// 		id := data["id"].(string)
// 		begin_time := utils.ToInt64(data["begin_time"])
// 		end_time := utils.ToInt64(data["end_time"])

// 		win_type := utils.ToInt64(data["win_type"])
// 		score := utils.ToInt64(data["score"])
// 		bet_amount := utils.ToInt64(data["bet_amount"])
// 		mines_multiple := utils.ToFloat64(data["mines_multiple"])

// 		beginTime := time.Unix(begin_time, 0).In(location)
// 		endTime := time.Unix(end_time, 0).In(location)

// 		record := &pb.FortuneGems2MyHistory{
// 			Id:       id,
// 			Date:     beginTime.Format(utils.FORMAT_DATE),
// 			Time:     beginTime.Format(utils.FORMAT_TIME),
// 			Edate:    endTime.Format(utils.FORMAT_DATE),
// 			Etime:    endTime.Format(utils.FORMAT_TIME),
// 			Bets:     bet_amount,
// 			Multiple: fmt.Sprintf("%.2f", mines_multiple),
// 			Score:    utils.CaseElse(score > 0, score+bet_amount, 0),
// 			WinType:  int32(win_type),
// 		}
// 		rsp.Records = append(rsp.Records, record)
// 	}
// 	rsp.HasMore = len(rsp.Records) >= int(arg.PageSize)
// }
