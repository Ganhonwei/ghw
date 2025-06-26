package fortune_gems2

import (
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/globalsign/mgo/bson"
)

func (a *DeskActor) Connected(ctx actor.Context) {
	msg := ctx.Message()
	//连接成功
	arg := msg.(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *DeskActor) Disconnected(ctx actor.Context) {
	msg := ctx.Message()
	//成功断开
	arg := msg.(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

func (a *DeskActor) CloseDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CloseDesk)
	glog.Debugf("CloseDesk %#v", arg)
	//移除
	delete(a.desks, arg.Roomid)
	delete(a.rules, arg.Unique)
	//TODO 优化,重复了
	//a.roomPid.Request(msg, ctx.Self())
	//响应
	//rsp := new(pb.ClosedDesk)
	//ctx.Respond(rsp)
}

func (a *DeskActor) LeaveDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.LeaveDesk)
	glog.Debugf("LeaveDesk %#v", arg)
	if v, ok := a.desks[arg.Roomid]; ok &&
		v.Number > 0 {
		v.Number--
		handler.RemoveRealUser(arg.Userid, v)
	}
	//响应
	//rsp := new(pb.LeftDesk)
	//ctx.Respond(rsp)
}

func (a *DeskActor) JoinDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.JoinDesk)
	glog.Debugf("JoinDesk %#v", arg)
	//房间数据变更
	if v, ok := a.desks[arg.Roomid]; ok {
		if !arg.Robot {
			v.RealUser = append(v.RealUser, arg.Userid)
		}
		v.Number++
	}
	//响应
	//rsp := new(pb.EnteredRoom)
	//ctx.Respond(rsp)
}

func (a *DeskActor) EnterDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.EnterDesk)
	// glog.Debugf("EnterDesk %#v", arg)
	a.enterDesk(arg, ctx)
}

func (a *DeskActor) SyncConfig(ctx actor.Context) {
	msg := ctx.Message()
	//同步配置
	arg := msg.(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	a.syncConfig(arg, ctx)
}

func (a *DeskActor) GetRoomList(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetRoomList)
	glog.Debugf("GetRoomList %#v", arg)
	if arg.Sender == nil {
		return
	}
	//响应
	rsp := handler.PackJHRoomList(arg, a.desks)
	arg.Sender.Tell(rsp)
}

func (a *DeskActor) ChangeDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangeDesk)
	glog.Debugf("ChangeDesk %#v", arg)
	// a.changeDesk(arg, ctx)
}

func (a *DeskActor) EarlyLeave(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.EarlyLeave)
	glog.Debugf("EarlyLeave %#v", arg)
	a.earlyLeave(arg, ctx)
}

func (a *DeskActor) ClearLeave(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ClearLeave)
	glog.Debugf("ClearLeave %#v", arg)
	a.clearLeave(arg, ctx)
}

func (a *DeskActor) GameStart(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GameStart)
	glog.Debugf("GameStart %#v", arg)
	a.gameStart(arg)
}

func (a *DeskActor) CloseServer(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CloseServer)
	glog.Debugf("CloseServer %#v", arg)
	a.closeServer = true
	a.broadcast(arg)
}

// '后台添加房间
func (a *DeskActor) syncConfig(arg *pb.SyncConfig, ctx actor.Context) {
	// handler.SyncConfig(arg, a.Name)
}

// 关闭一张桌子
func (a *DeskActor) _closeDesk(gameData *data.Game, ctx actor.Context) {
	glog.Debugf("close Desk %#v", gameData)
	//可以去room服务中取
	if k, ok := a.rules[gameData.Id]; ok {
		glog.Debugf("close Desk %s", k)
		a.stopDesk(k, ctx)
		delete(a.rules, gameData.Id)
	}
}

// 停止服务
func (a *DeskActor) stopDesk(roomid string, ctx actor.Context) {
	if v, ok := a.desks[roomid]; ok {
		//关闭房间消息
		msg1 := new(pb.ServeStop)
		v.Pid.Request(msg1, ctx.Self())
		//关闭房间消息
		//msg2 := new(pb.CloseDesk)
		//msg2.Roomid = roomid
		//a.roomPid.Request(msg2, ctx.Self())
		//delete(a.desks, roomid)
	}
}

// 是否跳过桌子
func (a *DeskActor) IsJumpOverRoom(roomId, userId string) bool {
	if v, ok := a.leave[roomId]; ok {
		for _, v := range v {
			if v == userId {
				return true
			}
		}
	}
	return false
}

//.

// '进入百人或者匹配房间
func (a *DeskActor) enterDesk(arg *pb.EnterDesk, ctx actor.Context) {

	rsp := new(pb.EnteredDesk)
	rsp.Gtype = arg.Gtype
	rsp.Rtype = arg.Rtype
	rsp.Gameid = arg.Gameid

	if a.closeServer {
		rsp.Error = pb.ServerClose
		ctx.Respond(rsp)
		return
	}

	user := new(data.User)
	err := json.Unmarshal(arg.Data, user)
	if err != nil {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	if arg.Roomid != "" {
		if d, ok := a.desks[arg.Roomid]; ok {
			d.Pid.Tell(arg)
			return
		} else {
			glog.Errorf("enter not exists roomid room: %s, %v", arg.Roomid, arg)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
	}

	//创建一个新的房间
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
	case int32(pb.ROOM_TYPE1): //私人
		glog.Errorf("enter Desk err %#v", arg)
	case int32(pb.ROOM_TYPE0): //自由
		// gameData := handler.NewHuaCoinGameData(arg.Gameid, desk_type)
		game := &data.Game{
			Id:       pb.GameType_name[int32(pb.FORTUNE_GEMS2)],
			Name:     pb.GameType_name[int32(pb.FORTUNE_GEMS2)],
			Gtype:    int32(pb.FORTUNE_GEMS2),
			Rtype:    int32(pb.ROOM_TYPE0),
			DeskType: int32(pb.DESK_TYPE_NORMAL),
		}
		game.Unique = data.ObjectIdString(bson.NewObjectId())
		if deskPid, ok := a.spawnDesk(game, ctx); ok {
			deskPid.Tell(arg)
			return
		}
	default:
	}
	rsp.Error = pb.Failed
	ctx.Respond(rsp)
	//arg.Sender.Tell(rsp)
}

func (a *DeskActor) earlyLeave(arg *pb.EarlyLeave, ctx actor.Context) {
	a.leave[arg.Roomid] = append(a.leave[arg.Roomid], arg.Userid)
}

// .清理离开数据
func (a *DeskActor) clearLeave(arg *pb.ClearLeave, ctx actor.Context) {
	if _, ok := a.leave[arg.Roomid]; ok {
		a.leave[arg.Roomid] = []string{}
	}
}

// 游戏开始
func (a *DeskActor) gameStart(arg *pb.GameStart) {
	if desk, ok := a.desks[arg.Rid]; ok {
		temp := desk.RealUser
		for _, v := range desk.RealUser {
			log := make(map[string]string)
			if _, ok := a.matchedLog[v]; !ok {
				a.matchedLog[v] = log
			} else {
				log = a.matchedLog[v]
			}
			for _, userid := range temp {
				if v == userid {
					continue
				}
				log[userid] = ""
			}
		}
	}
}

// '启动新服务,新开的房间同步状态
func (a *DeskActor) spawnDesk(gameData *data.Game,
	ctx actor.Context) (deskPid *actor.PID, ok bool) {
	deskData := handler.NewDeskData(gameData)
	return a.spawnDesk2(deskData, ctx)
}

func (a *DeskActor) spawnDesk2(deskData *data.DeskData,
	ctx actor.Context) (deskPid *actor.PID, ok bool) {
	glog.Debugf("spawn Desk start rid=%s, unique=%s, deskType=%d, rtype=%d", deskData.Rid, deskData.Unique, deskData.DeskType, deskData.Rtype)
	//新桌子
	newDesk1 := NewDesk(deskData)
	//spawn desk
	deskPid = newDesk1.newDesk()
	glog.Debugf("deskPid: %#v", deskPid.String())
	//添加桌子
	if !a.addDesk(deskData, deskPid, ctx) {
		//关闭房间消息
		msg1 := new(pb.ServeStop)
		deskPid.Request(msg1, ctx.Self())
		ok = false
		return
	}
	newDesk1.dbmsPid = a.dbmsPid
	newDesk1.roomPid = a.roomPid
	newDesk1.rolePid = a.rolePid
	// newDesk1.loggerPid = a.loggerPid
	newDesk1.selfPid = deskPid

	//添加新桌子
	a.desks[deskData.Rid] = &data.DeskBase{
		DeskData: deskData,
		Pid:      deskPid,
	}
	//规则暂时只关闭时用到
	a.rules[deskData.Unique] = deskData.Rid
	//启动
	deskPid.Tell(new(pb.ServeStart))
	glog.Debugf("spawn Desk successfully %s, %s",
		deskData.Rid, deskPid.String())
	// glog.Debugf("spawn Desk %#v", deskData)
	glog.Debugf("spawn Desk finish rid=%s, unique=%s, deskType=%d, rtype=%d", deskData.Rid, deskData.Unique, deskData.DeskType, deskData.Rtype)
	ok = true
	return
}

// 添加桌子
func (a *DeskActor) addDesk(deskData *data.DeskData,
	deskPid *actor.PID, ctx actor.Context) bool {
	//添加桌子到dbms
	msg2 := new(pb.AddDesk)
	msg2.Desk = deskPid
	msg2.Roomid = deskData.Rid
	msg2.Rtype = deskData.Rtype
	msg2.Gtype = deskData.Gtype
	msg2.Unique = deskData.Unique
	res2 := a.reqRoom(msg2, ctx)
	var response2 *pb.AddedDesk
	var ok bool
	if response2, ok = res2.(*pb.AddedDesk); !ok {
		glog.Errorf("add desk failed: %#v", res2)
		return false
	}
	if response2.Error != pb.OK {
		glog.Errorf("add desk failed: %v", response2.Error)
		return false
	}
	glog.Debugf("add Desk successfully %s, %s",
		response2.Roomid, response2.Code)
	deskData.Rid = response2.Roomid
	deskData.Code = response2.Code
	return true
}

// 登录成功数据处理
func (a *DeskActor) reqRoom(msg interface{}, ctx actor.Context) interface{} {
	timeout := 3 * time.Second
	res1, err1 := a.roomPid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRoom err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

// 广播消息
func (a *DeskActor) broadcast(msg interface{}) {
	for _, v := range a.desks {
		v.Pid.Tell(msg)
	}
}

// 匹配重置
func (a *DeskActor) resetMatching() {
	resetTime := utils.TimestampToday2(12, 0, 0)
	now := utils.LocalTime().Unix()
	if env == "dev" {
		// 测试环境6分钟刷新一次
		a.matchTimer++
		if a.matchTimer != 120 {
			return
		}
	} else {
		if now != resetTime {
			return
		}
	}
	a.matchTimer = 0
	glog.Infof("TP mathch reset time：%v", utils.LocalTime().Format("2006-01-02 15:04:05"))
	a.matchedLog = make(map[string]map[string]string)
}

// vim: set foldmethod=marker foldmarker=//',//.:
