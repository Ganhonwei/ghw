package dbms

import (
	"errors"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"sort"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 生成一个牌桌邀请码,全列表中唯一
func (a *RoomActor) genCode() (s string, err error) {
	for i := 0; i < 100_000; i++ { // 重复尝试
		s = utils.RandStr(6)
		//是否已经存在
		if _, ok := a.codes[s]; !ok {
			return
		}
	}
	err = errors.New("server room code fulled")
	return
}

//生成房间ID
//func (a *RoomActor) genDesk(arg *pb.GenDesk, ctx actor.Context) {
//	glog.Debugf("genDesk Rtype: %d, Gtype: %d", arg.Rtype, arg.Gtype)
//	rsp := new(pb.GenedDesk)
//	rsp.Roomid = a.uniqueid.GenID()
//	//TODO 百人,私人
//	//rsp.Code = a.genCode()
//	//响应
//	ctx.Respond(rsp)
//}

// 添加房间
func (a *RoomActor) AddDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.AddDesk)
	glog.Debugf("AddDesk: %v", arg)
	glog.Debugf("addDesk Rtype: %d, Gtype: %d", arg.Rtype, arg.Gtype)
	rsp := new(pb.AddedDesk)
	//已经存在
	if _, ok := a.rules[arg.Unique]; ok && arg.Unique != "" {
		glog.Errorf("addDesk err Rtype: %d, Gtype: %d ",
			arg.Rtype, arg.Gtype)
		glog.Errorf("addDesk err Roomid: %s, Unique: %s",
			arg.Roomid, arg.Unique)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	rsp.Roomid = a.uniqueid.GenID()
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE1): //私人
		//邀请码
		var err error
		rsp.Code, err = a.genCode()
		if err != nil {
			glog.Error("addDesk genCode err: ", err)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		//私人房间
		a.codes[rsp.Code] = &data.PrivRoom{
			Code:   rsp.Code,
			Roomid: rsp.Roomid,
			Gtype:  arg.Gtype,
			Rtype:  arg.Rtype,
		}
	}
	//响应消息
	ctx.Respond(rsp)
	//添加房间
	a.rooms[rsp.Roomid] = arg.Desk
	a.rules[arg.Unique] = rsp.Roomid
}

// 加入房间
func (a *RoomActor) JoinDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JoinDesk)
	glog.Debugf("JoinDesk %#v", arg)
	//房间数据变更
	if _, ok := a.router[arg.Userid]; !ok {
		a.router[arg.Userid] = arg.Roomid
		a.count[arg.Roomid]++
	}
	//响应
	//rsp := new(pb.JoinedDesk)
	//ctx.Respond(rsp)
}

// 离开房间
func (a *RoomActor) LeaveDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LeaveDesk)
	glog.Debugf("LeaveDesk %#v", arg)
	//移除
	if _, ok := a.router[arg.Userid]; ok {
		delete(a.router, arg.Userid)
		if n, ok := a.count[arg.Roomid]; ok && n > 0 {
			a.count[arg.Roomid] = n - 1
		}
	}
	//响应
	//rsp := new(pb.LeftDesk)
	//ctx.Respond(rsp)
}

func (a *RoomActor) Logout(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Logout)
	glog.Debugf("Logout %#v", arg)
	//TODO 暂时不处理
}

// 关闭房间
func (a *RoomActor) CloseDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CloseDesk)
	glog.Debugf("CloseDesk %#v", arg)
	//glog.Debugf("CloseDesk router %#v", a.router)
	//glog.Debugf("CloseDesk count %#v", a.count)
	//glog.Debugf("CloseDesk rules %#v", a.rules)
	delete(a.count, arg.Roomid)
	delete(a.codes, arg.Code)
	delete(a.rules, arg.Unique)
	delete(a.rooms, arg.Roomid)
	glog.Debugf("CloseDesk %d", len(a.rooms))
	//响应
	//rsp := new(pb.ClosedDesk)
	//ctx.Respond(rsp)
}

// 匹配房间
func (a *RoomActor) MatchDesk(ctx actor.Context) {
	msg := ctx.Message().(*pb.MatchDesk)
	glog.Debugf("matchDesk codes: %#v", a.codes)
	rsp := new(pb.MatchedDesk)
	rsp.Gameid = msg.Gameid
	rsp.Roomid = msg.Roomid
	rsp.Rtype = msg.Rtype
	rsp.Gtype = msg.Gtype
	rsp.Dtype = msg.Dtype
	rsp.Ltype = msg.Ltype
	if v, ok := a.codes[msg.Code]; ok &&
		msg.Code != "" {
		msg.Roomid = v.Roomid
	}
	if v, ok := a.rooms[msg.Roomid]; ok &&
		msg.Roomid != "" {
		rsp.Desk = v
		ctx.Respond(rsp)
		return
	}
	rsp.Error = pb.MatchFail
	ctx.Respond(rsp)
}

// 刷新房间人数
func (a *RoomActor) refreshRoomNum() {
	a.onlineNum = make(map[int32]map[string]int32)
	ntf := new(pb.GamePeopleNtf)
	peoples := config.GetRoomPeoples()
	// 当前时间段
	hour := utils.LocalTime().Hour()
	var interval []int32
	for _, p := range peoples {
		if 2 <= hour && hour < 10 {
			// 闲时
			interval = p.PeopleInterval1
		} else if 10 <= hour && hour < 18 {
			// 正常
			interval = p.PeopleInterval2
		} else {
			// 忙时
			interval = p.PeopleInterval3
		}
		if _, ok := a.onlineNum[p.Gtype]; !ok {
			a.onlineNum[p.Gtype] = make(map[string]int32)
		}
		if len(interval) <= 0 {
			continue
		}
		a.onlineNum[p.Gtype][p.RoomId] = utils.RandInt32N(interval[1]-interval[0]) + interval[0]
	}
	for k, v := range a.onlineNum {
		bean := new(pb.GamePeople)
		bean.Gtype = k
		for _, v2 := range v {
			bean.Count += v2
		}
		ntf.Game = append(ntf.Game, bean)
	}
	// 通知玩家
	nodePid.Tell(ntf)
}

func (a *RoomActor) SyncGamePeople(ctx actor.Context) {
	ntf := new(pb.GamePeopleNtf)
	for k, v := range a.onlineNum {
		bean := new(pb.GamePeople)
		bean.Gtype = k
		for _, v2 := range v {
			bean.Count += v2
		}
		ntf.Game = append(ntf.Game, bean)
	}
	ctx.Respond(ntf)
}

func (a *RoomActor) GetGameListReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetGameListReq)
	glog.Debugf("GetGameListReq %#v", arg)
	a.getGameListReq(arg, ctx)
}

// 获取游戏列表
func (a *RoomActor) getGameListReq(msg *pb.GetGameListReq, ctx actor.Context) {
	rsp := &pb.GetGameListRsp{}
	all_games := config.GetGames()
	var games data.GameList
	for _, v := range all_games {
		if msg.GameType != v.Gtype {
			continue
		}
		if v.Status == 0 {
			continue
		}
		// 跳过对战房
		if v.RoomType != 0 {
			continue
		}
		// if a.IsGameOpen(v.Gtype) {
		// 	games = append(games, v)
		// }
		games = append(games, v)
	}

	sort.Sort(games)
	for _, v := range games {
		info := &pb.GameListInfo{
			Id:        v.Id,
			GameType:  v.Gtype,
			MinAccess: int32(v.Min_Access),
			MaxAccess: int32(v.Max_Access),
			Sort:      int32(v.SortId),
			Bet:       v.GetBottom(),
			RoomType:  int32(v.RoomType),
		}
		if n, ok := a.onlineNum[msg.GameType]; ok {
			info.Online = n[v.Id]
		}
		rsp.List = append(rsp.List, info)
	}

	ctx.Respond(rsp)
}

// 打码量结算
func (a *RoomActor) ShareBetAmount(ctx actor.Context) {
	arg := ctx.Message().(*pb.ShareBetAmount)
	glog.Debugf("ShareBetAmount %#v", arg)
	share := &data.ShareBetAmount{
		Userid:   arg.Userid,
		Name:     arg.Name,
		Superior: arg.Superior,
		Score:    int64(math.Abs(float64(arg.Score))),
		Ctime:    utils.LocalTime().Unix(),
	}
	share.Save()
}

// 根据 code 进入私人房
func (a *RoomActor) PrivEnterRoomReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.PrivEnterRoomReq)
	glog.Debugf("PrivEnterRoomReq: %v", arg)
	rsp := new(pb.PrivEnterRoomRsp)

	privRoom, ok := a.codes[arg.Code]
	// data.PrivRoom
	if !ok {
		rsp.Error = pb.RoomNotExist
		ctx.Respond(rsp)
		return
	}
	roomPid, ok = a.rooms[privRoom.Roomid]
	if !ok {
		rsp.Error = pb.RoomNotExist
		ctx.Respond(rsp)
		return
	}

	// send to user
	rsp.Gtype = privRoom.Gtype
	rsp.Rtype = privRoom.Rtype
	rsp.Code = arg.Code

	// 判断前端游戏是否下载
	if len(arg.IsDownloadGame) > 0 {
		if down, ok := arg.IsDownloadGame[privRoom.Gtype]; ok && !down {
			rsp.Error = pb.GameNotDownload
			ctx.Respond(rsp)
			return
		}
	}

	ctx.Respond(rsp)

	msg := new(pb.MatchedDesk)
	msg.Roomid = privRoom.Roomid
	msg.Gtype = privRoom.Gtype
	msg.Rtype = privRoom.Rtype
	msg.Desk = roomPid
	// send to gate
	ctx.Sender().Tell(msg)
}

// 获取排行榜
// func (a *RoomActor) RankWithdrawReq(ctx actor.Context) {
// 	rsp := new(pb.RankWithdrawRsp)
// 	// today
// 	todayRanks, err := getRankWithdrawCurrent(context.Background(), rankWithdrawTodayKey, rankWithdrawLimit)
// 	if err != nil {
// 		rsp.Error = pb.Failed
// 		return
// 	}
// 	rsp.Today = mappingRankWithdraw(todayRanks)

// 	// this week
// 	weekRanks, err := getRankWithdrawCurrent(context.Background(), rankWithdrawWeekKey, rankWithdrawLimit)
// 	if err != nil {
// 		rsp.Error = pb.Failed
// 		return
// 	}
// 	rsp.Week = mappingRankWithdraw(weekRanks)

// 	// yesterday
// 	yesterday, err := getRankWithdrawExpired(context.Background(), rankWithdrawYesterdayKey)
// 	if err == nil {
// 		rsp.Yesterday = mappingRankWithdraw(yesterday)
// 	}

// 	// last week
// 	lastweek, err := getRankWithdrawExpired(context.Background(), rankWithdrawLastWeekdayKey)
// 	if err == nil {
// 		rsp.LastWeek = mappingRankWithdraw(lastweek)
// 	}
// 	ctx.Respond(rsp)
// }

func mappingRankWithdraw(users []*pb.EventRankWithdraw) (ranks []*pb.RankWithdraw) {
	for _, rank := range users {
		ranks = append(ranks, &pb.RankWithdraw{
			Userid:   rank.Userid,
			Username: rank.Username,
			Avatar:   rank.Avatar,
			Amount:   rank.Amount,
			VipLv:    rank.VipLv,
		})
	}
	return
}
