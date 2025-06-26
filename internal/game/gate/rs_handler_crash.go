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
func (rs *RoleActor) CRASHEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHEnterRoomReq)
	glog.Debugf("CRASHEnterRoomReq %#v", arg)
	msg := rs.enterCrashMatchDesk(ctx, arg.Roomid)
	if msg != nil {
		msg.Rtype = int32(pb.ROOM_TYPE2) //百人
		rs.selectDesk(msg, ctx)
	}
}

// 进入房间成功
func (rs *RoleActor) CRASHEnterSuccessReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHEnterSuccessReq)
	glog.Debugf("CRASHEnterSuccessReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.CRASHEnterSuccessRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 下注
func (rs *RoleActor) CRASHBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHBetReq)
	if rs.gamePid == nil {
		rsp := new(pb.CRASHBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.CRASHBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 取消下注
func (rs *RoleActor) CRASHCancelBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHCancelBetReq)
	if rs.gamePid == nil {
		rsp := new(pb.CRASHCancelBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	arg.Userid = rs.Userid
	rs.gamePid.Request(arg, ctx.Self())
}

// 撤离
func (rs *RoleActor) CRASHBackReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHBackReq)
	if rs.gamePid == nil {
		rsp := new(pb.CRASHBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 房间列表
func (rs *RoleActor) CRASHRoomListReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHRoomListReq)
	glog.Debugf("CRASHRoomListReq %#v", arg)
	rs.getCrashRoomList(arg, ctx)
}

// 设置自动撤离
func (rs *RoleActor) CRASHCrashMultipleReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHCrashMultipleReq)
	glog.Debugf("CRASHCrashMultipleReq %#v", arg)
	rs.autoCrashMultiple(arg, ctx)
}

// 离开
func (rs *RoleActor) CRASHLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CRASHLeaveReq)
	glog.Debugf("CRASHLeaveReq %#v", arg)
	if rs.gamePid == nil {
		rsp := new(pb.CRASHLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.gamePid.Request(arg, ctx.Self())
}

// 进入或匹配桌子
func (rs *RoleActor) enterCrashMatchDesk(ctx actor.Context, rid string) *pb.MatchDesk {
	//已经在游戏中,直接加入
	if rs.gamePid != nil {
		// game2 := config.GetGame(rs.gameId)
		// if game2.Gtype != int32(pb.CRASH) {
		if rs.gtype != int32(pb.CRASH) {
			rsp := new(pb.CRASHEnterRoomRsp)
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
	msg.Name = cfg.Section("game.crash").Name()
	msg.Gtype = int32(pb.CRASH) //crash
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	msg.Roomid = rid
	// if rs.PCSwitch {
	// 	msg.Dtype = int32(pb.DESK_TYPE_POINTCONTROL)
	// }
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) getCrashRoomList(arg *pb.CRASHRoomListReq, ctx actor.Context) {
	//获取游戏服节点或者房间进程
	msg := new(pb.GetRoomList)
	msg.Rtype = arg.Rtype
	msg.Userid = rs.User.GetUserid()
	msg.Sender = ctx.Self()
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.crash").Name()
	msg.Gtype = int32(pb.CRASH) //crash
	rs.dbmsPid.Request(msg, ctx.Self())
}

// 设置自动撤离
func (rs *RoleActor) autoCrashMultiple(arg *pb.CRASHCrashMultipleReq, ctx actor.Context) {
	user := rs.User
	user.CrashMultiple = arg.Multiple
	user.CrashAutoLeave = arg.AutoCrash
	rs.status = true
	rsp := &pb.CRASHCrashMultipleRsp{
		Multiple:  arg.Multiple,
		AutoCrash: arg.AutoCrash,
	}
	rs.Send(rsp)
	if rs.gamePid != nil {
		rs.gamePid.Request(arg, ctx.Self())
	}
}

// crash 个人记录查询
func (rs *RoleActor) CrashMyHistoryReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CrashMyHistoryReq)
	rsp := rs.crashMyHistoryReq(arg, int32(pb.CRASH))
	rs.Send(rsp)
}

// crash 个人记录查询
func (rs *RoleActor) crashMyHistoryReq(arg *pb.CrashMyHistoryReq, gtype int32) (rsp *pb.CrashMyHistoryRsp) {
	rsp = new(pb.CrashMyHistoryRsp)

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

	args := []any{userid, stime.Unix(), etime.Unix(), gtype}

	cond := ""
	if arg.PrevLastId != "" {
		cond += " AND begin_time < (SELECT begin_time FROM game.col_detail FINAL WHERE id = ? AND userid = ? AND gtype = ? LIMIT 1)"
		args = append(args, arg.PrevLastId, userid, gtype)
	}

	sql := `
		SELECT id, begin_time, win_type, crash_win_result2, score, bet_amount, 
			crash_multiple2, crash_multiple1_2, crash_bet0, crash_bet1, crash_win0, crash_win1
		FROM game.col_detail FINAL
		WHERE userid = ? AND begin_time >= ? AND begin_time < ? AND robot = 0 AND gtype = ? AND win_type in (1,2,3) %s
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
		win_type := utils.ToInt64(data["win_type"])
		crash_win_result2 := utils.ToInt64(data["crash_win_result2"])
		score := utils.ToInt64(data["score"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		crash_multiple2 := utils.ToInt64(data["crash_multiple2"])
		crash_multiple1_2 := utils.ToInt64(data["crash_multiple1_2"])
		crash_bet0 := utils.ToInt64(data["crash_bet0"])
		crash_bet1 := utils.ToInt64(data["crash_bet1"])
		crash_win0 := utils.ToInt64(data["crash_win0"])
		crash_win1 := utils.ToInt64(data["crash_win1"])

		beginTime := time.Unix(begin_time, 0).In(location)

		bets0 := utils.CaseElse(crash_bet1 == 0 && crash_bet0 == 0, bet_amount, crash_bet0) // 新字段老数据
		score0 := utils.CaseElse(crash_win1 == 0 && crash_win0 == 0, score, crash_win0)     // 新字段老数据
		winType0 := utils.CaseElse(score0 > 0, 1, utils.CaseElse(score0 < 0, 2, 3))
		winType1 := utils.CaseElse(crash_win1 > 0, 1, utils.CaseElse(crash_win1 < 0, 2, 3))
		record := &pb.CrashMyHistory{
			Id:        id,
			Date:      beginTime.Format(utils.FORMAT_DATE),
			Time:      beginTime.Format(utils.FORMAT_TIME_Minute),
			Bets:      bet_amount,
			Multiple:  int32(crash_win_result2),
			Score:     utils.CaseElse(score > 0, score+bet_amount, 0),
			Bets0:     bets0,
			Bets1:     crash_bet1,
			Multiple0: int32(crash_multiple2),
			Multiple1: int32(crash_multiple1_2),
			Score0:    utils.CaseElse(score0 > 0, score0+bets0, 0),
			Score1:    utils.CaseElse(crash_win1 > 0, crash_win1+crash_bet1, 0),
			WinType:   int32(win_type),
			WinType0:  int32(winType0),
			WinType1:  int32(winType1),
		}
		rsp.Records = append(rsp.Records, record)
	}
	rsp.HasMore = len(rsp.Records) >= int(arg.PageSize)
	return
}

// crash 大赢家记录查询
func (rs *RoleActor) CrashTopWinnersReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.CrashTopWinnersReq)
	rsp := rs.crashTopWinnersReq(arg, int32(pb.CRASH))
	rs.Send(rsp)
}

// crash 大赢家记录查询
func (rs *RoleActor) crashTopWinnersReq(arg *pb.CrashTopWinnersReq, gtype int32) (rsp *pb.CrashTopWinnersRsp) {
	rsp = new(pb.CrashTopWinnersRsp)

	now := time.Now().In(location)
	var stime time.Time
	switch arg.Range {
	case 2: // 月
		stime = now.AddDate(0, -1, 0)
	case 3: // 年
		stime = now.AddDate(-1, 0, 0)
	default: // 天
		stime = now.AddDate(0, 0, -1)
	}
	rsp.DateRange = fmt.Sprintf("%s to %s", stime.Format(utils.FORMAT), now.Format(utils.FORMAT))

	args := []any{stime.Unix(), gtype}
	sql := `
		SELECT (CASE WHEN crash_multiple2 > crash_multiple1_2 THEN crash_multiple2 ELSE crash_multiple1_2 END) big_multiple,
			(CASE WHEN crash_multiple2 > crash_multiple1_2 THEN 0 ELSE 1 END) big_pos,
			id, userid, robot, nickname, photo, vip_lv, begin_time, score, bet_amount, crash_win_result2, crash_bet0, crash_bet1, crash_win0, crash_win1
		FROM game.col_detail FINAL
		WHERE begin_time > ? AND gtype = ? AND (crash_multiple2 > 0 OR crash_multiple1_2 > 0)
		ORDER BY big_multiple DESC LIMIT 20
	`

	var datas []map[string]any
	err = ck.Select(&datas, sql, args...)
	if err != nil {
		rsp.Error = pb.Failed
		return
	}
	var baseUserids []string
	for _, data := range datas {
		id := data["id"].(string)
		userid := data["userid"].(string)
		nickname, _ := data["nickname"].(string)
		photo, _ := data["photo"].(string)
		vip_lv := utils.ToInt64(data["vip_lv"])
		begin_time := utils.ToInt64(data["begin_time"])
		crash_win_result2 := utils.ToInt64(data["crash_win_result2"])
		big_multiple := utils.ToInt64(data["big_multiple"])
		big_pos := utils.ToInt64(data["big_pos"])
		robot := utils.ToInt64(data["robot"])
		score := utils.ToInt64(data["score"])
		bet_amount := utils.ToInt64(data["bet_amount"])
		crash_bet0 := utils.ToInt64(data["crash_bet0"])
		crash_bet1 := utils.ToInt64(data["crash_bet1"])
		crash_win0 := utils.ToInt64(data["crash_win0"])
		crash_win1 := utils.ToInt64(data["crash_win1"])

		beginTime := time.Unix(begin_time, 0).In(location)
		bets0 := utils.CaseElse(crash_bet1 == 0 && crash_bet0 == 0, bet_amount, crash_bet0) // 新字段老数据
		score0 := utils.CaseElse(crash_win1 == 0 && crash_win0 == 0, score, crash_win0)     // 新字段老数据

		if robot == 0 && nickname == "" {
			baseUserids = append(baseUserids, userid)
		}

		record := &pb.CrashTopWinner{
			Id:           id,
			Date:         beginTime.Format(utils.FORMAT_DATE),
			Time:         beginTime.Format(utils.FORMAT_TIME_Minute),
			Userid:       userid,
			Nickname:     nickname,
			Photo:        photo,
			VipLv:        int32(vip_lv),
			Multiple:     int32(crash_win_result2),
			BackMultiple: int32(big_multiple),
			Bets:         utils.CaseElse(big_pos == 0, bets0, crash_bet1),
			Score:        utils.CaseElse(big_pos == 0, score0+bets0, crash_win1+crash_bet1),
		}
		rsp.Records = append(rsp.Records, record)
	}

	// 老数据个人信息补全
	if len(baseUserids) > 0 {
		var datas []map[string]any
		err = ck.Select(&datas, `
			SELECT userid, nickname, photo,vip_lv FROM game.col_user FINAL WHERE userid IN ?
		`, baseUserids)
		if err != nil {
			glog.Errorf("select user base info error: %v, %v", baseUserids, err)
		} else {
			for _, data := range datas {
				userid := data["userid"].(string)
				nickname := data["nickname"].(string)
				photo := data["photo"].(string)
				vip_lv := utils.ToInt64(data["vip_lv"])
				for _, record := range rsp.Records {
					if record.Userid == userid {
						record.Nickname = nickname
						record.Photo = photo
						record.VipLv = int32(vip_lv)
					}
				}
			}
		}
	}

	return
}
